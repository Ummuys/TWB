execPath=telegramBot

all: clean check_logs check_env build
	@./$(execPath)

build:
	@echo "Building . . .\n"
	@go build -o $(execPath)

check_logs:
	@echo "Check logs . . ."
	@if [ ! -d logs ]; then\
			echo "Creating 'logs/' folder . . .\n"; \
			mkdir logs; \
		else \
			echo "Logs folder exist!\n"; \
		fi

check_env:
	@echo "Check .env . . ."
	@if [ ! -f .env ]; then\
			echo "You didn't create .env! Do this!\n"; \
			exit 1; \
		else \
			echo ".env file exists!\n"; \
		fi

clean:
	@echo "Rebuilding!\n"
	@rm -rf $(execPath)
