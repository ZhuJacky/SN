.PHONY: backend checker notifier swagger deploy

# REGISTRY=harbor.trustasia.cn
REGISTRY?=harbor.trustasia.cn
PROJECT=mysslee
VERSION=`git describe --tags`

release:
	@standard-version

backend:
	@cd app/backend \
		&& go build -o backend \
		&& ./backend

checker:
	@cd app/checker \
		&& go build -o backend \
		&& ./backend

notifier:
	@cd app/notifier \
		&& go build -o backend \
		&& ./backend

_backend:
	@cd app/backend \
		&& CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -tags doc -o backend
	@docker build -t $(REGISTRY)/$(PROJECT)/qcloud-backend -f Dockerfile.backend .

_checker:
	@cd app/checker \
		&& CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -tags doc -o backend
	@docker build -t $(REGISTRY)/$(PROJECT)/qcloud-checker -f Dockerfile.checker .

_notifier:
	@cd app/notifier \
		&& CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -tags doc -o backend
	@docker build -t $(REGISTRY)/$(PROJECT)/qcloud-notifier -f Dockerfile.notifier .

_echoip:
	@cd app/echoip \
		&& CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o backend
	@docker build -t $(REGISTRY)/${PROJECT}/qcloud-echoip -f Dockerfile.echoip .

mock:
	@ cd test \
		&& GOOS=linux GOARCH=amd64 go build -o backend
	@docker build -t $(REGISTRY)/$(PROJECT)/qcloud-mock -f Dockerfile.mock . \
		&& docker push $(REGISTRY)/$(PROJECT)/qcloud-mock

dev_backend: _backend
	@docker tag $(REGISTRY)/$(PROJECT)/qcloud-backend $(REGISTRY)/$(PROJECT)/qcloud-backend:dev \
		&& docker push $(REGISTRY)/$(PROJECT)/qcloud-backend:dev

dev_checker: _checker
	@docker tag $(REGISTRY)/$(PROJECT)/qcloud-checker $(REGISTRY)/$(PROJECT)/qcloud-checker:dev \
		&& docker push $(REGISTRY)/$(PROJECT)/qcloud-checker:dev

dev_notifier: _notifier
	@docker tag $(REGISTRY)/$(PROJECT)/qcloud-notifier $(REGISTRY)/$(PROJECT)/qcloud-notifier:dev \
		&& docker push $(REGISTRY)/$(PROJECT)/qcloud-notifier:dev

dev_echoip: _echoip
	@docker tag $(REGISTRY)/$(PROJECT)/qcloud-echoip $(REGISTRY)/$(PROJECT)/qcloud-echoip:dev \
		&& docker push $(REGISTRY)/$(PROJECT)/qcloud-echoip:dev


tag_backend: _backend
	@docker tag $(REGISTRY)/$(PROJECT)/qcloud-backend $(REGISTRY)/$(PROJECT)/qcloud-backend:$(VERSION) \
		&& docker push $(REGISTRY)/$(PROJECT)/qcloud-backend:$(VERSION)

tag_checker: _checker
	@docker tag $(REGISTRY)/$(PROJECT)/qcloud-checker $(REGISTRY)/$(PROJECT)/qcloud-checker:$(VERSION) \
		&& docker push $(REGISTRY)/$(PROJECT)/qcloud-checker:$(VERSION)

tag_notifier: _notifier
	@docker tag $(REGISTRY)/$(PROJECT)/qcloud-notifier $(REGISTRY)/$(PROJECT)/qcloud-notifier:$(VERSION) \
		&& docker push $(REGISTRY)/$(PROJECT)/qcloud-notifier:$(VERSION)

tag:
	@rm deploy/v* && touch deploy/$(VERSION)

deploy: tag
	@cp -r conf deploy/services/
	@cd app/backend \
	  && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ../../deploy/services/backend
	@cd app/checker \
	  && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ../../deploy/services/checker
	@cd app/notifier \
	  && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ../../deploy/services/notifier
	@cd app/echoip \
	  && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ../../deploy/echoip/echoip

