

const TASK_PRIORITY = {
	HIGH: 1,
	MEDIUM: 2,
	LOW: 3	
};
      
const LOGGER_MODES = {
	// print logs more pretty
	PRETTY: 'pretty',
};

const LOGGER_NAMESPACES = {
	TASK: 'task',
	SERVER: 'server',
};

const CRON = {
	EVERY_MINUTE: '* * * * *',
	EVERY_TEN_SECONDS: '*/10 * * * * *'
};

module.exports = {
	LOGGER_NAMESPACES,
	LOGGER_MODES,
	CRON,
	TASK_PRIORITY,
};
