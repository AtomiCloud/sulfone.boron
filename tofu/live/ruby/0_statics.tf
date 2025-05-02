locals {

  cloudflare = {
    account_id        = "177aa484a66427793543c5e958f8d020"
    warp_policy_id    = "55971cbe-2ddf-4c30-b504-abbb901a0600"
    gateway_policy_id = "274146bb-6ccf-4988-bc61-79a070acbb1a"
    zone              = "atomi.cloud"
  }

  platforms = {

    sulfoxide = {
      name = "Sulfoxide"
      slug = "sulfoxide"

      services = {
        sos = {
          name        = "Secret of secrets"
          slug        = "sos"
          description = "Secrets of other infisical secrets"
        }
        metrics_engine = {
          name        = "Metrics Engine"
          slug        = "gallium"
          description = "Victoria Metrics to store and process metrics as the metrics engine"
        }
        terraform = {
          name        = "Terraform"
          slug        = "terraform"
          description = "Terraform"
        }
        tofu = {
          name        = "Tofu"
          slug        = "tofu"
          description = "IaC for AtomiCloud"
        }
        argocd = {
          name        = "ArgoCD"
          slug        = "helium"
          description = "Deployment platform using GitOps"
        }
        backup_engine = {
          name        = "Backup Engine"
          slug        = "fluorine"
          description = "Velero as the backup engine for Kubernetes"
        }
        secrets_engine = {
          name        = "Secrets Engine"
          slug        = "cobalt"
          description = "External Secrets to sync secrets from infisical"
        }
        internal_ingress = {
          name        = "Internal Ingress"
          slug        = "boron"
          description = "Internal Ingress with cloudflared"
        }
        external_ingress = {
          name        = "External Ingress"
          slug        = "gold"
          description = "External Ingress with nginx"
        }
        scaler = {
          name        = "Pod Autoscaler"
          slug        = "iron"
          description = "KEDA scaler for pods"
        }
        cluster_scaler = {
          name        = "Cluster Scaler"
          slug        = "krypton"
          description = "Karpenter scaler for nodes"
        }
        cluster_issuer = {
          name        = "Cluster Issuer"
          slug        = "zinc"
          description = "Cluster Issuer for Certificate"
        }
        cert_manager = {
          name        = "Cert Manager"
          slug        = "sulfur"
          description = "Certificate Issuing operator"
        }
        reloader = {
          name        = "Reloader"
          slug        = "chlorine"
          description = "Reloader"
        }
        policy_engine = {
          name        = "Policy Engine"
          slug        = "argon"
          description = "Kyverno operator"
        }
        policies = {
          name        = "Policies"
          slug        = "sodium"
          description = "Kyverno policies"
        }
        otel_collector = {
          name        = "OpenTelemetry Collector"
          slug        = "silicon"
          description = "OpenTelemetry Collector"
        }
        otel_operator = {
          name        = "OpenTelemetry Operator"
          slug        = "lithium"
          description = "OpenTelemetry Operator"
        }
        external_dns = {
          name        = "External DNS"
          slug        = "tin"
          description = "External DNS"
        }
        vcluster = {
          name        = "Virtual Cluster"
          slug        = "iodine"
          description = "Virtual clusters with vcluster"
        }
      }
    }
    
    sulfone = {
      name = "Sulfone"
      slug = "sulfone"
      services = {

        boron = {
          name        = "CyanPrint Executor"
          slug        = "boron"
          description = "Executor for CyanPrint to perform the actual templating"
        }

        zinc = {
          name        = "CyanPrint Registry"
          slug        = "zinc"
          description = "CyanPrint Registry for templates and processors"
        }

        argon = {
          name        = "CyanPrint Frontend"
          slug        = "argon"
          description = "CyanPrint Frontend for CyanPrint Registry"
        }
      }
    }

  }

  landscapes = {

    arceus = {
      name        = "Arceus"
      slug        = "arceus"
      description = "Global meta landscape-agnostic administrative environment"
    }

    pinsir = {
      name        = "Pinsir"
      slug        = "pinsir"
      description = "Continuous Integration"
    }

    lapras = {
      name        = "Lapras"
      slug        = "lapras"
      description = "Local development"
    }

    tauros = {
      name        = "Tauros"
      slug        = "tauros"
      description = "Local Development (Production Parity)"
    }

    absol = {
      name        = "Absol"
      slug        = "absol"
      description = "Test environment"

    }

    pichu = {
      name        = "Pichu"
      slug        = "pichu"
      description = "Singapore Development Environment"
    }

    pikachu = {
      name        = "Pikachu"
      slug        = "pikachu"
      description = "Singapore Staging Environment"
    }

    raichu = {
      name        = "Raichu"
      slug        = "raichu"
      description = "Singapore Production Environment"
    }

    suicune = {
      name        = "Suicune"
      slug        = "suicune"
      description = "Singapore administrative environment"
    }

    entei = {
      name        = "Entei"
      slug        = "entei"
      description = "Singapore physical landscape"
    }


  }

  clusters = {
    digital_ocean = {
      set   = "opal-ruby"
      blue  = "opal"
      green = "ruby"
    }
    linode = {
      set   = "onyx-jade"
      blue  = "onyx"
      green = "jade"
    }
    vultr = {
      set   = "mica-talc"
      blue  = "mica"
      green = "talc"
    }
    aws = {
      set   = "topaz-amber"
      blue  = "topaz"
      green = "amber"
    }
    gcp = {
      set   = "agate-lapis"
      blue  = "agate"
      green = "lapis"
    }
    azure = {
      set   = "beryl-coral"
      blue  = "beryl"
      green = "coral"
    }
  }
}