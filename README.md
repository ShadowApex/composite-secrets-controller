# Composite Secrets Controller

The Composite Secrets Controller is a Kubernetes controller that allows you
to combine the data from multiple ConfigMap and Secret sources into a single
Secret.

## Description

Very often configuration files defined as Secrets reference the same secrets
over and over again. The Composite Secrets Controller allows you to define
a single secret that combines the data from multiple ConfigMap or Secret
sources. This can also be especially useful in situations when you are storing
Kubernetes manifests in a Git repo and use [SOPS Encrypted Secrets](https://fluxcd.io/flux/guides/mozilla-sops/)
and want to store non-sensitive parts of a secret in plain-text and reference
encrypted secrets.

The controller works by watching for `CompositeSecret` objects that get created
in the cluster, and generates a corresponding secret that combines multiple
secrets into one.

An example of a `CompositeSecret` looks like this:

```yaml
apiVersion: composite.shadowblip.com/v1alpha1
kind: CompositeSecret
metadata:
  name: compositesecret-sample
spec:
  replacements:
    REPLACEME:
      secretRef:
        name: compositesecret-sample-source
        namespace: default
        key: mykey
  template:
    stringData:
      my-thing: |
        Here we say REPLACEME
```

There are two main parts to a `CompositeSecret`: replacements and template.
The "replacements" section defines any number of keywords you want to replace
in the "template" section along with a `secretRef` or `configMapRef` that the
controller will use to replace the keyword in the template with.

The "template" section contains your plain-text secret data. Any keywords found
in the data you define here will get replaced with the values defined in the
"replacements" section.

**NOTE: Make sure your keywords don't show up anywhere in your secret data or they will also be replaced!**

## Getting Started

You’ll need a Kubernetes cluster to run against. You can use [KIND](https://sigs.k8s.io/kind) to get a local cluster for testing, or run against a remote cluster.
**Note:** Your controller will automatically use the current context in your kubeconfig file (i.e. whatever cluster `kubectl cluster-info` shows).

### Running on the cluster

1. Install Instances of Custom Resources:

```sh
kubectl apply -f config/samples/
```

2. Build and push your image to the location specified by `IMG`:

```sh
make docker-build docker-push IMG=<some-registry>/composite-secrets-controller:tag
```

3. Deploy the controller to the cluster with the image specified by `IMG`:

```sh
make deploy IMG=<some-registry>/composite-secrets-controller:tag
```

### Uninstall CRDs

To delete the CRDs from the cluster:

```sh
make uninstall
```

### Undeploy controller

UnDeploy the controller to the cluster:

```sh
make undeploy
```

## Contributing

// TODO(user): Add detailed information on how you would like others to contribute to this project

### How it works

This project aims to follow the Kubernetes [Operator pattern](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/)

It uses [Controllers](https://kubernetes.io/docs/concepts/architecture/controller/)
which provides a reconcile function responsible for synchronizing resources untile the desired state is reached on the cluster

### Test It Out

1. Install the CRDs into the cluster:

```sh
make install
```

2. Run your controller (this will run in the foreground, so switch to a new terminal if you want to leave it running):

```sh
make run
```

**NOTE:** You can also run this in one step by running: `make install run`

### Modifying the API definitions

If you are editing the API definitions, generate the manifests such as CRs or CRDs using:

```sh
make manifests
```

**NOTE:** Run `make --help` for more information on all potential `make` targets

More information can be found via the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html)

## License

Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
