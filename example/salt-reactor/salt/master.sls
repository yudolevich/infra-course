gitea:
  pkg.installed:
  - name: docker.io

  pip.installed:
  - pkgs:
    - docker
    - pygit2

  docker_container.running:
  - image: gitea/gitea
  - port_bindings:
    - 3000:3000
