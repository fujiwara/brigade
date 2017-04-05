package brigade

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	stdlog "log"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/mitchellh/go-homedir"

	"golang.org/x/crypto/ssh"
	"gopkg.in/alessio/shellescape.v1"
)

const (
	NodeSeparator = "->"
	EOFMaker      = "--"
	LineDelimiter = "\n"
)

var (
	executable string
	hostname   string
	log        *stdlog.Logger
)

func init() {
	var err error

	hostname, err = os.Hostname()
	if err != nil {
		stdlog.Fatal("can't get hostname", err)
	}
	log = stdlog.New(os.Stderr, "["+hostname+"] ", stdlog.LstdFlags)
	log.Println("hostname", hostname)

	executable, err = os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("executable", executable)
}

type Delivery []string

func (d Delivery) Src() string {
	return d[0]
}

func (d Delivery) Dest() string {
	return d[1]
}

func (d Delivery) String() string {
	return strings.Join(d, NodeSeparator)
}

type Deliveries struct {
	ds []*Delivery
	wg sync.WaitGroup
}

func (ds Deliveries) String() string {
	b := bytes.NewBufferString("")
	for _, d := range ds.ds {
		io.WriteString(b, d.String())
		io.WriteString(b, LineDelimiter)
	}
	io.WriteString(b, EOFMaker)
	io.WriteString(b, LineDelimiter)
	return b.String()
}

func ParseLine(line string) (*Delivery, error) {
	if strings.HasPrefix(line, EOFMaker) {
		return nil, io.EOF
	}
	nodes := strings.Split(line, NodeSeparator)
	if len(nodes) != 2 {
		return nil, errors.New("invalid line")
	}
	return &Delivery{
		strings.TrimSpace(nodes[0]),
		strings.TrimSpace(nodes[1]),
	}, nil
}

func Parse(src io.Reader) (*Deliveries, io.Reader, error) {
	ds := Deliveries{
		ds: make([]*Delivery, 0),
	}
	r := bufio.NewReader(src)
	for {
		line, _, err := r.ReadLine()
		if err != nil {
			break
		}
		if len(line) == 0 {
			// first brank line found
			break
		}
		d, err := ParseLine(string(line))
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, nil, err
		}
		ds.ds = append(ds.ds, d)
	}
	return &ds, r, nil
}

func (ds *Deliveries) RemoteCommand(dst string, filename string, mode os.FileMode) {
	defer ds.wg.Done()
	log.Printf("%s -> %s:%v", hostname, dst, filename)

	f, err := os.Open(filename)
	if err != nil {
		log.Println("cannot open %s %s", filename, err)
		return
	}
	defer f.Close()

	key, err := ioutil.ReadFile(defaultIdentityFile())
	if err != nil {
		log.Printf("unable to read private key: %v", err)
		return
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		log.Printf("unable to parse private key: %v", err)
		return
	}

	clientConfig := &ssh.ClientConfig{
		User: os.Getenv("USER"),
		Auth: []ssh.AuthMethod{ssh.PublicKeys(signer)},
	}

	addr := fmt.Sprintf("%s:%d", dst, 22)
	conn, err := ssh.Dial("tcp", addr, clientConfig)
	if err != nil {
		log.Printf("cannot connect %v: %v", addr, err)
		return
	}
	defer conn.Close()

	session, err := conn.NewSession()
	if err != nil {
		log.Printf("cannot open new session: %v", err)
		return
	}
	defer session.Close()

	stdin, err := session.StdinPipe()
	if err != nil {
		log.Println(err)
		return
	}
	session.Stderr = os.Stderr
	session.Stdout = os.Stdout

	var wg2 sync.WaitGroup
	wg2.Add(1)
	go func() {
		defer wg2.Done()
		defer stdin.Close()
		if _, err := io.WriteString(stdin, ds.String()); err != nil {
			log.Println(err)
			return
		}
		if _, err := io.Copy(stdin, f); err != nil {
			log.Println(err)
			return
		}
	}()

	commandLine := []string{
		executable,
		"-mode", strconv.FormatInt(int64(mode), 8),
		shellescape.Quote(filename),
	}
	log.Printf("ssh %s %s", dst, strings.Join(commandLine, " "))
	if err := session.Run(strings.Join(commandLine, " ")); err != nil {
		log.Println(err)
	}
	wg2.Wait()
}

func Run(filename string, mode os.FileMode) {
	defer os.Stdin.Close()
	ds, r, err := Parse(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}
	if err := StoreFile(filename, r, mode); err != nil {
		log.Fatal(err)
	}
	for _, d := range ds.ds {
		if d.Src() != hostname {
			continue
		}
		ds.wg.Add(1)
		go ds.RemoteCommand(d.Dest(), filename, mode)
	}
	ds.wg.Wait()
}

func defaultIdentityFile() string {
	dir, err := homedir.Dir()
	if err != nil {
		log.Printf("could not get home dir: %v", err)
	}
	return dir + "/.ssh/id_rsa"
}

func StoreFile(name string, r io.Reader, mode os.FileMode) error {
	tmp, err := ioutil.TempFile(os.TempDir(), "stretcher")
	if err != nil {
		return err
	}
	defer tmp.Close()
	log.Printf("write file: %s", tmp.Name())

	if n, err := io.Copy(tmp, r); err != nil {
		return err
	} else {
		log.Printf("wrote %d bytes", n)
	}

	tmp.Close()
	log.Printf("rename %s to %s", tmp.Name(), name)
	if err := os.Rename(tmp.Name(), name); err != nil {
		return err
	}
	log.Printf("set mode %s", mode)
	return os.Chmod(name, mode)
}
