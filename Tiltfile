docker_build("hello-zele-operator", ".")

k8s_yaml("config/default")
k8s_resource("controller-manager", port_forwards=["8080:8080", "8081:8081"])
