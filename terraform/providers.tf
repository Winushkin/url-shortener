terraform {
  required_providers {
    kafka = {
      source  = "Mongey/kafka"
      version = "0.7.1" 
    }
  }
}

provider "kafka" {
  bootstrap_servers = ["localhost:9094"]
  
  # Оставляем только эти параметры для локального подключения без шифрования
  tls_enabled = false
  timeout     = 10
}
