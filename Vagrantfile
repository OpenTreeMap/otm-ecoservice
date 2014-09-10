# -*- mode: ruby -*-
# vi: set ft=ruby :

def local_ip
  `ipconfig getifaddr en0`.strip
end

VAGRANTFILE_API_VERSION = "2"

# Ensure role dependencies are in place
if [ "up", "provision" ].include?(ARGV.first) &&
  !(File.directory?("ansible/roles/azavea.golang") || File.symlink?("ansible/roles/azavea.golang"))

  unless system("ansible-galaxy install -r ansible/roles.txt -p ansible/roles")
    $stderr.puts "\nERROR: Please install Ansible 1.4.2+ so that the ansible-galaxy binary"
    $stderr.puts "is available."
    exit(1)
  end
end

Vagrant.configure(VAGRANTFILE_API_VERSION) do |config|
  config.vm.box = "ubuntu/trusty64"
  config.vm.hostname = "otm-ecoservice"

  # Wire up the proxy
  if Vagrant.has_plugin?("vagrant-proxyconf")
    config.proxy.http     = "http://#{local_ip}:8123/"
    config.proxy.https    = "http://#{local_ip}:8123/"
    config.proxy.no_proxy = "localhost,127.0.0.1"
  end

  # Mapping the local source code directory into the GOPATH inside the VM. Also, using azavea.com
  config.vm.synced_folder ".", "/home/vagrant/src/github.com/azavea/ecobenefits"

  config.vm.provision "ansible" do |ansible|
    ansible.playbook = "ansible/site.yml"
    ansible.sudo = true
    ansible.groups = {
      "development" => [ "default" ]
    }
  end
end
