
SENTRY_ORG=testorg-az
SENTRY_PROJECT=sentry-go-demo
GO_RELEASE_VERSION=`sentry-cli releases propose-version`

setup_release: create_release associate_commits

create_release:
	sentry-cli releases -o $(SENTRY_ORG) new -p $(SENTRY_PROJECT) $(GO_RELEASE_VERSION)

associate_commits:
	sentry-cli releases -o $(SENTRY_ORG) -p $(SENTRY_PROJECT) set-commits --auto $(GO_RELEASE_VERSION)

deploy: setup_release build run

build:
	go build

run:
	./sentry-go-demo

