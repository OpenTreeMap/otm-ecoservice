# -*- mode: ruby -*-
# vi: set ft=ruby :

Vagrant.require_version ">= 1.5"

VAGRANT_ROOT = File.dirname(__FILE__)

# Uses the contents of roles.txt to ensure that ansible-galaxy is run if any
# dependencies are missing.
def install_dependent_roles
  File.foreach(File.join(VAGRANT_ROOT, "ansible/roles.txt")) do |line|
    role_path = File.join(VAGRANT_ROOT, "ansible/roles/#{line.split(",").first}")

    if !File.directory?(role_path) && !File.symlink?(role_path)
      unless system("ansible-galaxy install -f -r "\
        "#{File.join(VAGRANT_ROOT, "ansible/roles.txt")} -p #{File.dirname(role_path)}")
        $stderr.puts "\nERROR: An attempt to install Ansible role dependencies failed."
        exit(1)
      end

      break
    end
  end
end

# Install missing role dependencies based on the contents of roles.txt
if [ "up", "provision" ].include?(ARGV.first)
  install_dependent_roles
end

VAGRANTFILE_API_VERSION = "2"
VAGRANT_PROXYCONF_ENDPOINT = ENV["VAGRANT_PROXYCONF_ENDPOINT"]

Vagrant.configure(VAGRANTFILE_API_VERSION) do |config|
  config.vm.box = "ubuntu/trusty64"
  config.vm.hostname = "otm-ecoservice"

  # Wire up the proxy if:
  #
  #   - The vagrant-proxyconf Vagrant plugin is installed
  #   - The user set the VAGRANT_PROXYCONF_ENDPOINT environmental variable
  #
  if Vagrant.has_plugin?("vagrant-proxyconf") &&
     !VAGRANT_PROXYCONF_ENDPOINT.nil?
    config.proxy.http     = VAGRANT_PROXYCONF_ENDPOINT
    config.proxy.https    = VAGRANT_PROXYCONF_ENDPOINT
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
