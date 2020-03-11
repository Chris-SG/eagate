#!/bin/bash

sha=$(git ls-remote git://github.com/chris-sg/eagate_models.git HEAD | awk '{ print $1}')
go get github.com/chris-sg/eagate_models@$sha