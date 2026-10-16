# Azure CSI driver for azurefile


See more: https://github.com/kubernetes-sigs/azurefile-csi-driver/tree/master/charts

```shell
make update
```

## Using The Addon

You need to replace the following values with the actual ones:

* `INSTALL_AZNFS_MOUNT` can be used to define the value of `INSTALL_AZNFS_MOUNT` env variable which controls if aznfs apt package should be installed on the nodes or not.
  * Possible values are `"true"`or `"false"`
  * Default is `"true"`.
