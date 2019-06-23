# How KubeOne uses SSH
As KubeOne requires ssh access to control plane nodes, somehow ssh
public/private keys should be handled. KubeOne doesn't handle decryption of
private SSH keys but instead rely on `ssh-agent`. In most cases we recommend
using `ssh-agent` as the easiest way to have your SSH keys *encrypted* at rest
and still useful for KubeOne.

## ssh-agent
If your operating system of choice doesn't do this for you automatically, you
can use something like
```bash
eval `ssh-agent`
```

and then later
```bash
ssh-add ~/.ssh/my_cool_custom_private_key
```

in order to cache it in ssh-agent memory for later use.

KubeOne is able to contact ssh-agent via socket (environment variable
`SSH_AUTH_SOCK`) and ask for authentication without getting unencrypted private
key.

## Providing SSH private keys directly, without ssh-agent
In rare case when it's not possible to use `ssh-agent`, you can provide private
key directly to KubeOne. The caveat is that private SSH key should be
unencrypted and thus we DON'T recommend this.

### Option 1, config manifest
You can point KubeOne to the unencrypted private SSH key using config API.

```yaml
hosts:
- publicAddress: '1.2.3.4'
  ...
  sshPrivateKeyFile: '/home/me/.ssh/my_cleantext_private_key'
```

### Option2, terraform output
You can also provide unencrypted private SSH key using terraform integration.

```terraform
output "kubeone_hosts" {
  value = {
    control_plane = {
      public_address       = my_vm_provider_server.control_plane.*.ipv4_address
      ...
      ssh_private_key_file = "/home/me/.ssh/my_cleantext_private_key"
    }
  }
}
```

## gpg-agent and ssh
It's possible to use GnuPG agent (`gpg-agent`) in replace of `ssh-agent`. It has
number of advantages, but it's also more complicated to setup.

In your `.bash_profile`
```bash
export SSH_AUTH_SOCK=$(gpgconf --list-dirs agent-ssh-socket)
gpgconf --launch gpg-agent
```

See more info about how to setup your SSH keys in GnuPG:
* https://opensource.com/article/19/4/gpg-subkeys-ssh
* https://opensource.com/article/19/4/gpg-subkeys-ssh-multiples
