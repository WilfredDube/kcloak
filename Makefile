lint:
	@$(shell go env GOPATH)/bin/golangci-lint run

test:
	./run-tests.sh

start-keycloak: stop-keycloak
	docker compose up -d

stop-keycloak:
	docker compose down

generate-kcloak-interface:
	@echo "Remember to: go install github.com/vburenin/ifacemaker@latest"
	@$(shell go env GOPATH)/bin/ifacemaker -f client.go -s KCloak -i KCloakIface -p kcloak -o kcloak_iface.go
