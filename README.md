zwo
===
Version 2

## Provisioner
There are different provisioners depending on what kind of system should be provisioned. Currently supported are SSH and
Docker. The following subsections will discuss certain requirements and constraints.

### SSH
Some debian systems don't have a root account you can directly log on too, but require to use sudo. This is kind of
tedious in some situations and makes matters difficult for server provisioning. But there is one benefit: while there
are constant attacks against certain well known account names (like root, www, news, ...) normal user accounts are
highly unlikely to be hit by a request. Together with only allowing ssh to certain accounts (like the set of
administrators) and only allowing ssh login via public key authentication the system should be pretty secure even with
ssh being available.

There are some requirements though. All administrators must be in the sudoers group and sudo should be
configured to not
ask for passwords. This can be done editing the sudoers file (use the `sudo visudo` command to get there). The following
line should be modified:
   %sudo ALL=(ALL:ALL) ALL
   to be like
      %sudo ALL=(ALL:ALL) NOPASSWD:ALL
