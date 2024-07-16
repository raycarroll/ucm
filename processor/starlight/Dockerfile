FROM ubuntu:focal
MAINTAINER Cristina Catalan <cataln.cris@gmail.com>

RUN pwd
WORKDIR /docker/starlight/

COPY ./STARLIGHTv04/* /docker/starlight/STARLIGHTv04/

#COPY ./config_files_starlight/StCv04.C11.arp220.config /docker/starlight/STARLIGHTv04/
COPY ./config_files_starlight/mask/* /docker/starlight/shared_directory/config_files_starlight/mask/

#COPY ./config_files_starlight/grid_example.in /docker/starlight/shared_directory/config_files_starlight/
COPY ./config_files_starlight/spectrum/* /docker/starlight/shared_directory/config_files_starlight/spectrum/
#RUN mkdir -p //starlight/shared_directory/output

RUN ls
RUN pwd

RUN echo "deb [trusted=yes] http://old-releases.ubuntu.com/ubuntu/ hardy universe" >> /etc/apt/sources.list
RUN echo "deb-src [trusted=yes] http://old-releases.ubuntu.com/ubuntu/ hardy universe" >> /etc/apt/sources.list
RUN echo "deb [trusted=yes] http://old-releases.ubuntu.com/ubuntu/ hardy-updates universe" >> /etc/apt/sources.list
RUN echo "deb-src [trusted=yes] http://old-releases.ubuntu.com/ubuntu/ hardy-updates universe" >> /etc/apt/sources.list

RUN apt update -y
RUN apt install -y gfortran
