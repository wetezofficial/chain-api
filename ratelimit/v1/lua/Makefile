.PHONY: test
test:
	rm -rf luacov.stats.out ; busted -c test-ratelimit.lua

.PHONY: coverage
coverage: test
	rm -rf luacov.report.out ; luacov '^ratelimit' && cat luacov.report.out
