* With the release of version 1.x.x we are releasing a lot of new features, please checkout the [changelog]()

* To migrate to the new version, please use the [migration script](https://github.com/codefresh-io/venona/blob/master/scripts/migration.sh)

* We do not expect any unexpected behaviour for users that are already running previous versions ( `version < 1.0.0` )

* Installation with previous `venona` binray will require now to speficy the exact version, `venona install ... --venona-version 0.30.2` otherwise, you will get the most recent version, which will not competible with the previous installation flow.
