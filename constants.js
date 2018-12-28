const AGENT_MODES = {
	// The agent will run inside a cluster as a pod
	IN_CLUSTER: 'InCluster',
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
	AGENT_MODES,
	LOGGER_NAMESPACES,
	LOGGER_MODES,
	CRON,
};
