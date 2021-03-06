APP_NAME:=uploadcdn
APP_PATH:=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
SCRIPT_PATH:=$(APP_PATH)/scripts
COMPILE_OUT:=$(APP_PATH)/bin/$(APP_NAME)
APP_CLIENT_NAME:=upload-client
COMPILE_CLIENT_OUT:=$(APP_PATH)/bin/$(APP_CLIENT_NAME)


build:
	@echo ">>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>making $@<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<"
	@chmod +x $(SCRIPT_PATH)/build/*.sh
	@cd $(APP_PATH)/cmd/server && $(SCRIPT_PATH)/build/gobuild.sh $(APP_NAME) $(COMPILE_OUT)

build-client:
	@echo ">>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>making $@<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<"
	@chmod +x $(SCRIPT_PATH)/build/*.sh
	@cd $(APP_PATH)/cmd/client && $(SCRIPT_PATH)/build/gobuild.sh $(APP_CLIENT_NAME) $(COMPILE_CLIENT_OUT)


linuxbuild-client:
	@echo ">>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>making $@<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<"
	@chmod +x $(SCRIPT_PATH)/build/*.sh
	@cd $(APP_PATH)/cmd/client && $(SCRIPT_PATH)/build/linuxbuild.sh $(APP_CLIENT_NAME) $(COMPILE_CLIENT_OUT)



server:export EGO_DEBUG=true
server:
	@echo ">>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>making $@<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<"
	@cd $(APP_PATH) && go run cmd/server/main.go --config=config/local.toml

client:export EGO_DEBUG=true
client:
	@echo ">>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>making $@<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<"
	@cd $(APP_PATH) && go run cmd/client/main.go --dir="./data"


server-test:export EGO_DEBUG=true
server-test:
	@echo ">>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>making $@<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<"
	@cd $(APP_PATH) && go run cmd/server/main.go --config=config/test.toml
