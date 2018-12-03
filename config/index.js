// eslint-disable-next-line import/no-unresolved
const { version } = require('./../package.json');
const { LOGGER_MODES } = require('./../constants');

function build() {
	return {
		metadata: {
			name: process.env.AGENT_NAME,
			version,
			mode: process.env.AGENT_MODE,
		},
		logger: {
			prettyPrint: !(process.env.LOGGER_MODE === LOGGER_MODES.PRETTY),
			level: process.env.LOGGER_LEVEL || 'info',
		},
		server: {
			port: process.env.PORT || '9000',
		},
		kubernetes: {
			config: {
				url: process.env.KUBERNETES_HOST,
				auth: {
					bearer: Buffer.from(process.env.KUBERNETES_AUTH_BEARER_TOKEN, 'base64'),
				},
				ca: Buffer.from(process.env.KUBERNETES_CA_CERT, 'base64'),
			},
		},
		codefresh: {
			baseURL: process.env.CODEFRESH_HOST || 'https://g.codefresh.io',
			token: process.env.CODEFRESH_TOKEN,
		},
		tasks: {
			FetchTasksToExecute: {
				cronExpression: process.env.TASK_FETCH_JOBS_TO_EXECUTE_CRON_EXPRESSION || '* * * * *', // once a minute
			},
			ReportStatus: {
				cronExpression: process.env.TASK_REPORT_STATUS_CRON_EXPRESSION || '* * * * *', // twice a minute
			},
		},
	};
}

module.exports = build;
