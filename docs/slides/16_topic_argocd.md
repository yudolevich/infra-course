## ArgoCD

```{image} ../img/argocd.png
:width: 200px
```

### Concepts
```{revealjs-fragments}
* Application
* Target State
* Live State
* Sync
* Health
```

### Architecture

```{image} ../img/argocd-slides1.png
:width: 600px
```

### Application

```{revealjs-code-block} yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: guestbook
  namespace: argocd
spec:
  project: default
  source:
    repoURL: https://github.com/argoproj/argocd-example-apps.git
    targetRevision: HEAD
    path: guestbook
  destination:
    server: https://kubernetes.default.svc
    namespace: guestbook
```

### Tools
```{revealjs-fragments}
* directory(yaml/json/jsonnet)
* kustomize
* helm
* custom plugin
```

### ApplicationSet

```{revealjs-code-block} yaml
---
data-line-numbers: 1-4|5-16|17-28
---
apiVersion: argoproj.io/v1alpha1
kind: ApplicationSet
metadata:
  name: guestbook
spec:
  goTemplate: true
  goTemplateOptions: ["missingkey=error"]
  generators:
  - list:
      elements:
      - cluster: engineering-dev
        url: https://1.2.3.4
      - cluster: engineering-prod
        url: https://2.4.6.8
      - cluster: finance-preprod
        url: https://9.8.7.6
  template:
    metadata:
      name: '{{.cluster}}-guestbook'
    spec:
      project: my-project
      source:
        repoURL: https://github.com/infra-team/cluster-deployments.git
        targetRevision: HEAD
        path: guestbook/{{.cluster}}
      destination:
        server: '{{.url}}'
        namespace: guestbook
```

### UI

```{image} ../img/argocd-ui.gif
:width: 700px
```
