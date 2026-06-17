terraform {
  required_version = ">= 1.3.0"

  required_providers {
    yandex = {
      source  = "yandex-cloud/yandex"
      version = ">= 0.100.0"
    }
    kafka = {
      source  = "Mongey/kafka"
      version = "~> 0.7.0"
    }
  }

  # Хранение состояния в S3-бакете Yandex Cloud
  backend "s3" {
    endpoints = { s3 = "https://storage.yandexcloud.net" }
    bucket   = "yc-url-shortener-tfstate" 
    key      = "prod/terraform.tfstate"   
    region   = "ru-central1"

    skip_region_validation      = true
    skip_credentials_validation = true
    skip_requesting_account_id  = true 
  }
}

# Инициализация провайдера Yandex Cloud
provider "yandex" {}
