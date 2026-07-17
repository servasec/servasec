.PHONY: all build-target build-scanners scan update-fixtures clean help

TEST_TARGETS_DIR ?= ../../servasec-test-targets
PARSER_TESTDATA  := ../backend/parsers/testdata

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}'

all: build-target build-scanners scan ## Build everything + run all scanners
	@echo "Run 'make update-fixtures' to copy reports to parser testdata"

build-target: ## Build the vulnerable test target image
	docker build -t servasec-test-target $(TEST_TARGETS_DIR)
	@echo "Saving image as tar..."
	docker save servasec-test-target -o /tmp/servasec-test-target.tar

build-scanners: ## Build the scanner tools image
	docker build -t servasec-scanners -f Dockerfile.scanners .

scan: build-target build-scanners ## Run all scanners against the test target
	@echo "Running scanners..."
	docker run --rm \
		-v /tmp/servasec-test-target.tar:/target/image.tar:ro \
		-v $(TEST_TARGETS_DIR):/target/files:ro \
		-v $(PWD)/reports:/reports \
		servasec-scanners
	@echo ""
	@echo "Reports generated in test/reports/"

update-fixtures: scan ## Copy scanner reports to parser testdata
	@mkdir -p $(PARSER_TESTDATA)
	cp reports/grype.json $(PARSER_TESTDATA)/grype_real.json
	cp reports/trivy.json $(PARSER_TESTDATA)/trivy_real.json
	cp reports/trivy-iac.json $(PARSER_TESTDATA)/trivy_iac_real.json
	cp reports/bandit.json $(PARSER_TESTDATA)/bandit_real.json
	cp reports/gosec.json $(PARSER_TESTDATA)/gosec_real.json
	cp reports/checkov.json $(PARSER_TESTDATA)/checkov_real.json
	cp reports/tfsec.json $(PARSER_TESTDATA)/tfsec_real.json
	cp reports/gitleaks.json $(PARSER_TESTDATA)/gitleaks_real.json
	cp reports/semgrep.json $(PARSER_TESTDATA)/semgrep_real.json
	cp reports/nuclei.json $(PARSER_TESTDATA)/nuclei_real.json
	cp reports/osv-scanner.json $(PARSER_TESTDATA)/osv-scanner_real.json
	@echo "Fixtures updated in $(PARSER_TESTDATA)/"
	@echo "Add test functions in your *_test.go files to parse these."

clean: ## Remove generated reports and temp files
	rm -f /tmp/servasec-test-target.tar
	rm -rf reports/*.json
	docker rmi servasec-test-target 2>/dev/null || true
	docker rmi servasec-scanners 2>/dev/null || true
