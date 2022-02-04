.PHONY: all
all: get-new-abusers ipset-server unban-ip-abuser

DEPENDS=internal/app/configs/*.go internal/app/utils/*.go

get-new-abusers: cmd/get-new-abusers/get-new-abusers.go $(DEPENDS)
	go build $<
ipset-server: cmd/ipset-server/ipset-server.go $(DEPENDS)
	go build $<
unban-ip-abuser: cmd/unban-ip-abuser/unban-ip-abuser.go $(DEPENDS)
	go build $<