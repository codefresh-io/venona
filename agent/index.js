
const Promise = require('bluebird');
const scheduler = require('node-schedule');
const Codefresh = require('./../services/Codefresh');
const Kubernetes = require('./../services/Kubernetes');
const Logger = require('./../services/Logger');
const { Server } = require('./../server');
const FetchTasksToExecute = require('./../tasks/FetchTasksToExecute');
const ReportStatus = require('./../tasks/ReportStatus');
const { LOGGER_NAMESPACES, AGENT_MODES } = require('./../constants');
const utils = require('./../utils');

const ERROR_MESSAGES = {
	FAILED_TO_CONSTRUCT_ANONYMOUS_AGENT: 'Cannot construct anonymous agent',
};

class Agent {
	constructor(config = {}) {
		this.name = utils.getPropertyOrError(config, 'metadata.name', ERROR_MESSAGES.FAILED_TO_CONSTRUCT_ANONYMOUS_AGENT);
		this.logger = Logger.create(config.metadata, config.logger);
		this.logger.info('Starting agent');
		this.server = new Server(config.metadata, config.server, this.logger.child({
			namespace: LOGGER_NAMESPACES.SERVER,
		}));
		this.codefreshAPI = new Codefresh(config.metadata, config.codefresh);
		this.kubernetesAPI = config.metadata.mode === AGENT_MODES.IN_CLUSTER
			? Kubernetes.buildFromInCluster(config.metadata)
			: Kubernetes.buildFromConfig(config.metadata, config.kubernetes);
		this.tasks = config.tasks;
	}

	async init() {
		this.logger.info('Initializing agent');
		return Promise.all([
			this.server.init(),
			this.codefreshAPI.init(),
			this.kubernetesAPI.init(),
		])
			.then(() => {
				this.logger.info('Initializing finished');
			}, (err) => {
				const message = `Failed to initialize agent with error message: ${err.message}`;
				this.logger.error(message);
				throw new Error(message);
			});
	}

	async start() {
		this._startTask(this.tasks.FetchTasksToExecute.cronExpression, FetchTasksToExecute);
		this._startTask(this.tasks.ReportStatus.cronExpression, ReportStatus);
		return Promise.resolve();
	}

	_startTask(cron, Task) {
		this.logger.info(`Starting task: ${Task.name} with cron: ${cron}`);
		scheduler.scheduleJob(cron, () => {
			const taskLogger = this.logger.child({
				namespace: LOGGER_NAMESPACES.TASK,
				task: FetchTasksToExecute.name,
			});
			Promise.resolve()
				.then(() => new Task(this.codefreshAPI, this.kubernetesAPI, taskLogger).run())
				.catch((err) => {
					this.logger.error(`Failed to execute task ${Task.name} with error message: ${err.message}`);
				})
				.done();
		});
	}
}

Agent.Errors = ERROR_MESSAGES;

module.exports = Agent;
