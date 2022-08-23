

once: install_kubebuilder


GOPATH = $(HOME)/go
PATH :=$(PATH):$(GOPATH)/bin:$(HOME)/.local/bin
DOMAIN = mytest.io
PROJECT = testoperator
OS := $(shell go env GOOS)
ARCH := $(shell go env GOARCH)

install_kubebuilder:
# download kubebuilder and install locally.
	$(info "Downloading from"  https://go.kubebuilder.io/dl/latest/$(OS)/$(ARCH) )
	$(shell curl -L -o kubebuilder  https://go.kubebuilder.io/dl/latest/$(OS)/$(ARCH) )
	$(shell chmod +x kubebuilder && mv kubebuilder $(HOME)/.local/bin)

clean:
	rm -rf $(PROJECT)
init_project:
# initialize the project.
	$(info "Initializing project" $(PROJECT))
	-$(shell mkdir ./$(PROJECT) )
	$(shell cd $(PROJECT) && kubebuilder init project --domain=$(DOMAIN) --repo $(DOMAIN)/$(PROJECT))

create_crd:
	cd $(PROJECT) && kubebuilder create api --group grpcapp --version v1 --kind Testoperartor && make manifests
