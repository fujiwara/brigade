FROM alpine:latest

RUN apk --no-cache add openssh
RUN for t in rsa dsa ecdsa ed25519; do ssh-keygen -t $t -f /etc/ssh/ssh_host_${t}_key -P ""; done

RUN adduser -h /home/test -D test
USER test
RUN ssh-keygen -f /home/test/.ssh/id_rsa -P "" && cat /home/test/.ssh/id_rsa.pub >> /home/test/.ssh/authorized_keys && chmod 600 /home/test/.ssh/authorized_keys

USER root
RUN passwd -u test

CMD /usr/sbin/sshd -D
