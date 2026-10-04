# MindCare proto toolchain. Requires buf >= 1.32 (https://buf.build/docs/installation).
# Generation is intentionally NOT part of the build yet.

.PHONY: lint format breaking generate clean

lint:
	buf lint

format:
	buf format -w

breaking:
	buf breaking --against '../../../.git#branch=main,subdir=app/backend/proto'

generate:
	buf generate

clean:
	rm -rf gen
