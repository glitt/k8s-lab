# For setting these 3 variables use:
# [VERSION=v3] [REGISTRY="gcr.io"] [PROJECT_ID="your_project_id"] gnumake release-image
VERSION?=v1
REGISTRY?=gcr.io
PROJECT_ID?=your_project_id

release-image: clean-image build-image push-image clean-image

# Builds a docker image that builds the app and packages it into a minimal docker image.
build-image:
	docker build --build-arg version=$(VERSION) -t $(REGISTRY)/$(PROJECT_ID)/kbmd:$(VERSION) .

# Push the image to an registry.
push-image:
	docker tag $(REGISTRY)/$(PROJECT_ID)/kbmd:$(VERSION) $(REGISTRY)/$(PROJECT_ID)/kbmd:$(VERSION)
	docker push $(REGISTRY)/$(PROJECT_ID)/kbmd:$(VERSION)

# Remove previous images and containers.
clean-image:
	docker rm -f $(REGISTRY)/$(PROJECT_ID)/kbmd:$(VERSION) 2> /dev/null || true

# Assemble everything.
pray-everything-works:
	kubectl apply -f cluster/kbmd-letsencrypt-issuer-prod.json
	kubectl apply -f cluster/kbmd-certificate.json
	kubectl apply -f cluster/kbmd-deployment.json
	kubectl apply -f cluster/kbmd-service.json
	kubectl apply -f cluster/kbmd-ingress.json
	kubectl apply -f cluster/kbmd-tls-ingress.json

# Tear down everything.
teardown:
	kubectl delete -f cluster/kbmd-ingress.json
	kubectl delete -f cluster/kbmd-service.json
	kubectl delete -f cluster/kbmd-deployment.json
	kubectl delete -f cluster/kbmd-certificate.json
	kubectl delete -f cluster/kbmd-letsencrypt-issuer-prod.json

.PHONY: release-image clean-image build-image push-image pray teardown
