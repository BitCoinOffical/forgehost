SERVICES := auth-service hosting-service notification-service social-service

.PHONY: test test-%
.PHONY: lint lint-%

test:
	@for s in $(SERVICES); do $(MAKE) test-$$s || exit 1; done

test-%:
	cd backend/$* && go test -v -race ./...

lint:
	@for s in $(SERVICES); do $(MAKE) lint-$$s || exit 1; done

lint-%:
	cd backend/$* && golangci-lint run ./...