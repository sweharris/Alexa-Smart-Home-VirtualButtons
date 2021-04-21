SRC = $(wildcard *.go)

GATEWAY=$(shell cat GATEWAY 2>/dev/null)
PASS=$(shell cat PASS 2>/dev/null)

main: $(SRC)
	CGO_ENABLED=0 go build -trimpath -o main $(SRC)
	strip main

lambda.zip: main
	rm -f lambda.zip
	zip lambda.zip main

deploy: lambda.zip
	aws lambda update-function-code --function-name Smart_Home_Virtual_Buttons --zip-file fileb://lambda.zip

# This should cause an error
test: main
	curl $(GATEWAY)

# This will trigger button1
push1: main
	curl $(GATEWAY) -H "Authorization: $(PASS)" -d '{"command": "pushcontact", "param1": "1"}'

pushname: main
	curl $(GATEWAY) -H "Authorization: $(PASS)" -d '{"command": "pushcontactbyname", "param1": "Test Button 1"}'

getbuttons: main
	curl $(GATEWAY) -H "Authorization: $(PASS)" -d '{"command": "getbuttons"}'

badpswd: main
	curl $(GATEWAY) -H "Authorization: A bad password" -d '{"command": "pushcontactbyname", "param1": "Test Button 1"}'
