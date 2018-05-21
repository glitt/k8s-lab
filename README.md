# K8S Lab

Deploy a TLS enabled demo app on GKE.

## Foreword

This, being a very simple lab excercise, is not intended for production use even though it can
easily be modified to suit production purposes. This guide does not explain how to set up GKE and
install necessary tooling, since it is not in the scope of the guide.

## Prerequisites

This guide assumes that you have a Google Cloud account, project and Kubernetes cluster already set
up with GKE. You can set up cluster by clicking through the GUI and then copy for command line for
future reference. `gcloud` should also be set up, your project should already be set up, Docker is
properly installed and configured with Google Cloud Container Registry.

### GNU Make

`gnumake` is used throughout the document, if `make` points to the GNU version of the tool, use that
(not the case on BSD and Darwin). This guide was tested only on latest OSX.

### Helm

Install helm by going to [](https://github.com/kubernetes/helm) and follow the guide.

### Service Account

Create service account for tiller: `kubectl create serviceaccount -n kube-system tiller`.

Create `ClusterRoleBinding`:

```
kubectl create clusterrolebinding tiller-binding --clusterrole=cluster-admin --serviceaccount \
    kube-system:tiller
```

Install tiller:
`helm init --service-account tiller`

Once tiller pod becomes ready, update chart repositories: `helm repo update`.

### Cert Manager

![Masterful diagram isn't it](https://github.com/glitt/k8s-lab/raw/master/cert-manager.png)

`cert-manager` is used for provisioning Let's Encrypt certificates:

```
helm install --name cert-manager \
    --namespace kube-system stable/cert-manager
```

Set your email in `cluster/kbmd-letsencrypt-issuer-prod.json` and
`cluster/kbmd-letsencrypt-issuer-staging.json`.

Use `cluster/kbmd-letsencrypt-issuer-staging.json` when testing, since rate limits are more relaxed,
but for production use, revert back to `letsencrypt-prod` issuer. Remember to change manifest in
`cluster/kbmd-certificate.json` to use `letsencrypt-staging` for testing purposes and adjust
`Makefile` accordingly.

When using `letsencrypt-staging`, the cert issued is signed by "Fake LE Intermediate X1", signed by
root CA which is not in client trust stores.

Hint: since JSON is used, manifests can easily be scripted to set these parameters up.

Learn more about Cert Manager [here](https://github.com/jetstack/cert-manager/).

### IP address

Reserve global IP address with: `gcloud compute addresses create kbmd-ip --global`. Point your
domains A record to this IP address. In `tools` directory there is an example for Google Cloud Dns.

## Bulding and releasing Docker image

There is a convenient `Makefile` for doing this. Release Docker images with:

```
VERSION=v1 PROJECT_ID=your_project_id gnumake release-image
```

which in turn will invoke `clean-image`, `build-image`, `push-image`, `clean-image` targets. You can
also modify `Makefile` to include your `PROJECT_ID`.

Essential part of building Docker image is testing the code. Code is built and tested using official
`golang` image as builder image and then binary in next stage is "copied" from `builder` stage. This
way resulting container is as minimal as possible and includes only the app. Using multistage builds
provides easier debugging, making specific stages like `testing` with different test battery,
building binaries with debugging symbols first and then production builds later, etc. Currently,
debugging information is stripped from resulting binary, making the image even smaller. If image
size is a concern (currently sitting at around 4.5 MiBs), it can be shrunk down further with tools
like `upx` and included in multistage build.

Underlying Go app is set up so that `VERSION` is being embedded into a variable `version` when
linked. After this is done, your image should be ready to be served from registry that you chose
(default is `gcr.io`).

## Deploying cluster

Now, here is where it gets fun. There is a single command that you need to execute in order to bring
the whole cluster up: `gnumake pray-everything-works` (sigh). This will (hopefully) bootstrap entire
cluster properly.

### Architecture

Basic diagram:

![Basic overview](https://github.com/glitt/k8s-lab/raw/master/basic-architecture.png)

### Bootstrapping

First, extensions should be configured. Apply `kbmd-letsencrypt-issuer-prod.json` with
`kubectl apply -f cluster/kbmd-letsencrypt-issuer-prod.json` manifest. This will create a "Cluster
Issuer" resource which will be later references by "Certificate" resource for information how to
obtain a certificate. "Cluster Issuer" represents a certificate authority which is used to obtain
certificates, while "Certificate" resource should contain all metadata required for issuing a cert.

Then, apply `kbmd-certificate.json` with: `kubectl apply -f cluster/kbmd-certificate.json`. This
will create "Certificate" resource.

Then, apply `kbmd-deployment.json` with `kubectl apply -f cluster/kbmd-deployment.json`. This will
create a deployment.

After deployment is created, create "Service" resource:`kubectl apply -f cluster/kbmd-service.json`.
To tie everything together, create "Ingres" resource, by applying both `cluster/kbmd-ingress.json`
and `cluster/kbmd-tls-ingress.json`.

### Destroying

Just like bootstrapping, just reversed. Use `gnumake teardown` to tear down. Execute `gcloud
clusters delete --region REGION_NAME CLUSTER_NAME` to delete entire cluster.

## Future work

There are many shortcomings to this approach, to name some:
  - Follow security best practices (RBAC, Security Context, logging, etc.).
  - Cert Manager is not exactly the best solution (long provisioning times, flaky at times).
  - Prevent AB-BA errors during rollouts.
  - No monitoring whatsoever.
  - Better documentation and diagrams :-)

Hope it was fun! :-)
