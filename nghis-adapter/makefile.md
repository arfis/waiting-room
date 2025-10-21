[comment]: <> (generated file do not edit)
# Makefile goals:

- download-all - downloads all dependencies and tools for the project
- run-dev - runs project with dev configuration
- run-prod - runs project with prod configuration
- build - builds binary
- build-image - builds docker image
- lint - runs linter for the project

## Database
to create new migration run
```
make create-migration
```
and you will be prompted to input migration name. New migration wil appear in ./migrations folder, and you can write your up and down sql scripts

to migrate database up run
```
make migrate-up
```
to migrate test database up run
```
make migrate-up-test
```

## Testing
All tests must be in a file with postfix \*\_test. All test functions in this file have to start with Test_ prefix and consume t *testing.T like so
```
Test_myFunction(t *testing.T){
  assert.Equal(t, 1, 1)
}
```

to run all tests run
```
make test-all
```

## Benchmarks
All benchmarks must be in a file with postfix \*\_test. All benchmark functions in this file have to start with Benchmark_ prefix and consume b *testing.B. Functions have to contain a for loop with iterations over b.N. Benchmark will iterate over that function until it will get an average run value.
```
Benchmark_myFunction(b *testing.B){
    initData()
	for n := 0; n<b.N;n++ {
	    a := doSomethingThatWillBeBenchmarked()
        assert.Equal(t, 1, a)
    }
}
```