#!/usr/bin/env bash

gometalinter --install

echo "Checking gometalinter..."

CONCURRENCY=${CONCURRENCY:-8}

gometalinter \
	--deadline=300s \
	--concurrency=$CONCURRENCY \
	--skip=vendor \
	--exclude="should have comment or be unexported" \
	./...

if [ $? -eq 1 ]; then
    exit 1
fi

EXIT_CODE=0

echo "Checking gofmt..."

for file in `ls |grep -v vendor | xargs -I {} find {} -name "*.go"|xargs -I {} gofmt -l {}`; do
	echo "need to format $file"
	EXIT_CODE=1
done

if [ $EXIT_CODE == 1 ]; then
  exit 1
fi