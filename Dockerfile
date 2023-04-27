FROM ruby:3.2.2

ARG RUBYGEMS_VERSION=3.3.20

RUN mkdir /myapp
WORKDIR /myapp

ADD Gemfile /myapp/Gemfile
ADD Gemfile.lock /myapp/Gemfile.lock

RUN bundle install
ADD . /myapp
