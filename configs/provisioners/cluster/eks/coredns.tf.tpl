locals {
  coredns_scheduling_patch = {
    spec = {
      template = {
        spec = {
          nodeSelector = {{ if hasKeyAny .distribution "nodeSelector" }}{
            {{- range $key, $value := .distribution.nodeSelector }}
            "{{ $key }}" = "{{ $value }}"
            {{- end }}
          } {{ else }} null {{ end }}
          tolerations = [
            {{- range $key, $value := .distribution.tolerations }}
            {
              key = "{{ $value.key }}"
              value = "{{ $value.value }}"
              effect = "{{ $value.effect }}"
            },
            {{- end }}
          ]
        }
      }
    }
  }
 coredns_scheduling_patch_as_json = jsonencode(local.coredns_scheduling_patch)
}

resource "local_file" "cluster_ca" {
   depends_on = [
     module.fury
   ]

  content = base64decode(data.aws_eks_cluster.fury.certificate_authority.0.data)
  filename = "${path.module}/secrets/ca.crt"
}

resource "null_resource" "patch_coredns" {
  depends_on = [
    module.fury,
    local_file.cluster_ca
  ]

  triggers = {
    run_once = local.coredns_scheduling_patch_as_json
  }

  provisioner "local-exec" {
    command = <<-EOT
      ${var.kubectl_path} patch deployment/coredns -n kube-system -p '${local.coredns_scheduling_patch_as_json}' \
      --server=${data.aws_eks_cluster.fury.endpoint} \
      --token=${data.aws_eks_cluster_auth.fury.token} \
      --certificate-authority=${path.module}/secrets/ca.crt \
      EOT
  }
}
