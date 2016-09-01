FROM golang

ADD . /go/src/github.com/PrairieLearn/autograd

RUN echo deb http://httpredir.debian.org/debian testing main > /etc/apt/sources.list.d/testing.list
RUN echo "Package: *\n\
Pin: release a=stable\n\
Pin-Priority: 900" > /etc/apt/preferences.d/stable.pref
RUN echo "Package: *\n\
Pin: release a=testing\n\
Pin-Priority: 750" > /etc/apt/preferences.d/testing.pref

RUN apt-get update
RUN apt-get install -y pkg-config
RUN apt-get -yt testing install libgit2-dev

RUN go install github.com/PrairieLearn/autograd/cmd/autograd

ENV AUTOGRAD_ROOT=/opt/autograd
RUN mkdir $AUTOGRAD_ROOT

ENTRYPOINT /go/bin/autograd
