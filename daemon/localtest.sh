#!/bin/bash

go build -o jsonexporter . || exit

./jsonexporter --config ../testdata/config.yaml
