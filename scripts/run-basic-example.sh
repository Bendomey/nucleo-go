
#!/bin/sh
srcPath="examples"
exampleFile="basic.go"
app="nucleo-basic-example"
src="$srcPath/$app/$exampleFile"

printf "\nStart running: $app\n"
time go run $src
printf "\nStopped running: $app\n\n"