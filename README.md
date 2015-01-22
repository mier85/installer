installer
=========

Just a simple installer, as for the moment, just for init.d scripts and without many configuration options.

Thanks to [https://github.com/fhd/init-script-template] for the template!

installation
============
go get github.com/mier85/installer

usage
=====
just import the package and call "installer.Register()"

todo
====
 - better error handling 
 - rollback for incomplete installations
 - uninstall
 - support more than just init.d scripts


