SRC = $(wildcard *.go)

GATEWAY=$(shell cat GATEWAY 2>/dev/null)
PASS=$(shell cat PASS 2>/dev/null)

bootstrap: $(SRC)
	CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -trimpath -o bootstrap $(SRC)
	strip bootstrap

lambda.zip: bootstrap
	rm -f lambda.zip
	zip lambda.zip bootstrap

deploy: lambda.zip
	aws lambda update-function-code --function-name Smart_Home_Virtual_Buttons --zip-file fileb://lambda.zip

# This should cause an error
test: bootstrap
	curl $(GATEWAY)

# This will trigger button1
push1: bootstrap
	curl $(GATEWAY) -H "Authorization: $(PASS)" -d '{"command": "pushcontact", "param1": "1"}'

pushname: bootstrap
	curl $(GATEWAY) -H "Authorization: $(PASS)" -d '{"command": "pushcontactbyname", "param1": "Test Button 1"}'

getbuttons: bootstrap
	curl $(GATEWAY) -H "Authorization: $(PASS)" -d '{"command": "getbuttons"}'

badpswd: bootstrap
	curl $(GATEWAY) -H "Authorization: A bad password" -d '{"command": "pushcontactbyname", "param1": "Test Button 1"}'

insecure: bootstrap
	curl "$(GATEWAY)?cmd=$(PASS)/pushcontact/1"
