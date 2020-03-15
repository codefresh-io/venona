// eslint-disable-next-line import/no-unresolved
const { version } = require('./../package.json');
const { CRON } = require('./../constants');

function build() {
	return {
		metadata: {
			version,
			id: process.env.AGENT_ID,
			venonaConfDir: process.env.VENONA_CONFIG_DIR,
		},
		logger: {
			...(!process.env.LOGGER_MODE && {
				prettyPrint: {
					levelFirst: true,
				}
			}),
			level: process.env.LOGGER_LEVEL || 'info',
		},
		server: {
			port: process.env.PORT || '9000',
		},
		kubernetes: {
			metadata: {
				name: process.env.SELF_POD_NAME,
				namespace: process.env.SELF_POD_NAMESPACE
			}
		},
		codefresh: {
			baseURL: process.env.CODEFRESH_HOST || 'https://g.codefresh.io',
			token: process.env.CODEFRESH_TOKEN,
		},
		jobs: {
			TaskPullerJob: {
				cronExpression: process.env.JOB_PULL_TASKS_TO_EXECUTE_CRON_EXPRESSION || CRON.EVERY_TEN_SECONDS,
			},
			StatusReporterJob: {
				cronExpression: process.env.JOB_REPORT_STATUS_CRON_EXPRESSION || CRON.EVERY_MINUTE,
				runOnce: true,
			},
			DEFAULT_CRON: CRON.EVERY_MINUTE,
			queue: {
				concurrency: parseInt(process.env.JOBS_QUEUE_CONCURRENCY || '1')
			}
		},
	};
}

module.exports = build;
