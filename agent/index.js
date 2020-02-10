const Promise = require('bluebird');
const Chance = require('chance');
const scheduler = require('node-schedule');
const path = require('path');
const _ = require('lodash');
const Queue = require('async/priorityQueue');
const recursive = require('recursive-readdir');
const Codefresh = require('./../services/Codefresh');
const Kubernetes = require('./../services/Kubernetes');
const Logger = require('./../services/Logger');
const { Server } = require('./../server');
const { LOGGER_NAMESPACES, AGENT_MODES } = require('./../constants');


const ERROR_MESSAGES = {
};

class Agent {

	_onEmptyQueue() {
		this.logger.info('Queue is empty');
	}

	_queueRunner(job = { run: Promise }, cb) {
		this.logger.info(`Running job: ${job.constructor.name}`);
		Promise.resolve()
			.then(() => job.exec())
			.then(() => cb(), cb);
	}

	async _loadJobs() {
		const ignorePaths = [(file, stats) => {

			return !(new RegExp(/.*job.js/g).test(file)) && !stats.isDirectory();
		}];
		return Promise
			.fromCallback(cb => recursive(path.join(__dirname, './../jobs'), ignorePaths, cb))
			.map(require)
			.map(Job => this._startJob(Job));
	}



	async init(config = {}) {
		try {
			// this.logger.info('Initializing agent');
			this.logger = Logger.create(config.metadata, config.logger);
			this.logger.info('Starting agent');
			this.server = new Server(config.metadata, config.server, this.logger.child({
				namespace: LOGGER_NAMESPACES.SERVER,
			}));
			this.codefreshAPI = new Codefresh(config.metadata, config.codefresh);
			this.kubernetesAPI = config.metadata.mode === AGENT_MODES.IN_CLUSTER
				? await Kubernetes.buildFromInCluster(config.metadata)
				: Kubernetes.buildFromConfig(config.metadata, config.kubernetes);
			this.jobs = config.jobs;
			this.queue = Queue(this._queueRunner.bind(this), config.jobs.queue.concurrency);
			this.queue.drain = this._onEmptyQueue.bind(this);
			await Promise.all([
				this.server.init(),
				this.codefreshAPI.init(),
				this.kubernetesAPI.init(),
			]);
			this.logger.info('All services has been initialized');
			await this._loadJobs();
		} catch(err) {
			const message = `Failed to initialize agent with error message: ${err.message}`;
			this.logger.error(message);
			throw new Error(message);
		}
	}

	_startJob(Job) {
		const cron = _.get(this, `jobs.${Job.name}.cronExpression`, this.jobs.DEFAULT_CRON);
		this.logger.info(`Preparing job: ${Job.name} with cron: ${cron}`);
		scheduler.scheduleJob(cron, () => {
			const taskLogger = this.logger.child({
				namespace: LOGGER_NAMESPACES.TASK,
				job: Job.name,
				uid: new Chance().guid(),
			});
			const job = new Job(this.codefreshAPI, this.kubernetesAPI, taskLogger);
			this.logger.info(`Pushing job: ${Job.name} to queue`);
			this.queue.push(job, 1, this._handleJobError(job));
		});
	}

	_handleJobError(job) {
		return (err) => {
			if (err) {
				this.logger.error(`Failed to execute job ${job.constructor.name} with error message: ${err.message}`);
				this.logger.error(err.stack);
			} else {
				this.logger.info(`Job: ${job.constructor.name} finished`);
			}
		};
	}
}

Agent.Errors = ERROR_MESSAGES;

module.exports = Agent;
