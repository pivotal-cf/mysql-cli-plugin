resource "tls_private_key" "provisioning-ssh-key" {
  algorithm = "RSA"
  rsa_bits  = 4096
}

resource "google_compute_firewall" "firewall-allow-external-access-to-ssh-and-mysql" {
  name    = "percona-pam-dare-test-firewall-access-rules"
  network = "default"

  allow {
    protocol = "tcp"
    ports    = ["22", "3306"]
  }

  source_ranges = ["0.0.0.0/0"]
  target_tags   = ["external-access"]
}

resource "google_compute_instance" "percona-pam-and-dare" {
  name = "percona-pam-dare-1"

  tags = ["percona-pam-and-dare", "external-access"]

  machine_type = "n1-standard-1"

  zone = "us-central1-a"

  boot_disk {
    initialize_params {
      image = "ubuntu-1804-lts"
    }
  }

  scratch_disk {
    interface = "SCSI"
  }

  network_interface {
    network = "default"

    access_config {
      // Ephemeral IP
    }
  }

  provisioner "remote-exec" {
    connection {
      type = "ssh"
      user = "root"
      host = self.network_interface[0].access_config[0].nat_ip
      timeout = "600s"
      private_key = tls_private_key.provisioning-ssh-key.private_key_pem
    }

    script = "setup.sh"
  }

  metadata = {
    ssh-keys = "root:${tls_private_key.provisioning-ssh-key.public_key_openssh}"
  }

  depends_on = [google_compute_firewall.firewall-allow-external-access-to-ssh-and-mysql]

}

output "ssh_private_key" {
  value = tls_private_key.provisioning-ssh-key.private_key_pem
  sensitive   = true
}

output "instance_ip" {
  value = google_compute_instance.percona-pam-and-dare.network_interface[0].access_config[0].nat_ip
}
