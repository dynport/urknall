# Urknall - opinionated provisioning for clever developers

[![Documentation](http://img.shields.io/badge/gh--pages-documentation-blue.svg)](http://urknall.dynport.de)
[![GoDoc](https://godoc.org/github.com/dynport/urknall?status.svg)](https://godoc.org/github.com/dynport/urknall)

[Urknall](http://urknall.dynport.de) is the basic building block for creating
go based tools for the administration of complex infrastructure. It provides
the mechanisms to access resources and keep a cache of executed tasks.
Description of tasks is done using a mix of explicit shell commands and
abstractions for more complex actions being pragmatic, but readable.

* _Automate provisioning_: Urknall provides automated
  provisioning, so that resources can be set up reproducibly and easily.
  Thereby Urknall helps with scaling infrastructure to the required size.
* _Agentless tool that only relies on common UNIX tools and provides decent
  caching_: As Urknall works agentless on most UNIX based systems, adding new
  resources to the infrastructure is effortless. The caching keeps track of
  performed tasks and makes repeated provisioning with thoughtful changes
  possible.
* _Template mechanism that helps at the start, but get’s out of the way later
  on_: Urknall provides some basic templates, but lets users modify those or
  add new ones to solve the specific problem at hand.

urknall is developed and maintained by [Dynport GmbH](http://www.dynport.de),
the company behind the translation management software
[PhraseApp](https://phraseapp.com).

