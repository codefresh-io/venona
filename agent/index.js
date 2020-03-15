const Promise = require('bluebird');
const Chance = require('chance');
const scheduler = require('node-schedule');
const path = require('path');
const _ = require('lodash');
const fs = require('fs');
const yaml = require('js-yaml');
const Queue = require('async/priorityQueue');
const recursive = require('recursive-readdir');
const Codefresh = require('./../services/Codefresh');
const Kubernetes = require('./../services/Kubernetes');
const Logger = require('./../services/Logger');
const { Server } = require('./../server');
const { LOGGER_NAMESPACES } = require('./../constants');


const ERROR_MESSAGES = {};

class Agent {

	async init(config = {}) {
		try {
			this.logger = Logger.create(config.metadata, config.logger);
			this.logger.info('Initializing agent');
			this.server = new Server(config.metadata, config.server, this.logger.child({
				namespace: LOGGER_NAMESPACES.SERVER,
			}));
			this.codefreshAPI = new Codefresh(config.metadata, config.codefresh);
			this.jobs = config.jobs;
			this.queue = Queue(this._queueRunner.bind(this), config.jobs.queue.concurrency);
			this.queue.drain = this._onEmptyQueue.bind(this);
			await Promise.all([
				this.server.init(),
				this.codefreshAPI.init(),
				this._prepareRuntimes(config),
			]);
			this.logger.info('All services has been initialized');
			await this._loadJobs();
		} catch(err) {
			const message = `Failed to initialize agent with error, message: ${err.message}`;
			this.logger.error(message);
			throw new Error(message);
		}
	}

	async _prepareRuntimes(config = {}) {
		this.logger.info(`Reading Venona config file directory: ${config.metadata.venonaConfDir}`);
		const cnf = await this._readFromVenonaConfDir(config.metadata.venonaConfDir);
		this.runtimes = {};
		_.map(cnf, (runtimecnf) => {
			if (this.runtimes[runtimecnf.name]) {
				this.logger.warn(`Ignoring runtime "${runtimecnf.name}", already been configured, conflict with file: "${runtimecnf.file}"`);
				return;
			}
			this.runtimes[runtimecnf.name] = {
				spec: _.omit(runtimecnf, 'name'),
			};
		});
		await Promise.all(_.map(this.runtimes, async (runtime, name) => {
			this.logger.info(`Initializing Kubernetes client for runtime: ${name}`);
			const opt = {
				config: {
					url: runtime.spec.host,
					auth: {
						bearer: runtime.spec.token
					},
					ca: runtime.spec.crt
				},
			};
			const client = Kubernetes.buildFromConfig(config.metadata, opt);
			try {
				await client.init();
				runtime.kubernetesAPI = client;
				this.logger.info(`Runtime ${name} was loaded successfuly`);
			} catch(err) {
				this.logger.error(`Failed to initiate runtime: ${name}, error: ${err.message}`);
				runtime.metadata.error = err.message;
			}
		}));
	}

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
			.map(Job => {
				const runOnce = _.get(this, `jobs.${Job.name}.runOnce`, false);
				if (runOnce) {
					this._runOnce(Job);
				}
				return Job;
			}).map(Job => this._startJob(Job));
			
	}

	async _readFromVenonaConfDir(dir) {
		const runtimeRegexp = /.*\.runtime\.yaml/;
		const data = await Promise.fromCallback(cb => fs.readdir(dir, cb));
		const runtimes = await Promise.all(_.map(data, async(f) => {
			const fullPath = path.join(dir, f);
			const stat = await Promise.fromCallback(cb => fs.stat(fullPath, cb));
			if (stat.isDirectory()) {
				this.logger.warn(`Directory "${f}" ignored, Venona loading only files that are mached to regexp ${runtimeRegexp.toString()}`);
				return Promise.resolve();
			}
			if (runtimeRegexp.test(f)) {
				await Promise.fromCallback(cb => fs.access(fullPath, cb));
				const venonaConf = await Promise.fromCallback(cb => fs.readFile(fullPath, cb));
				return {
					...yaml.safeLoad(venonaConf.toString()),
					file: fullPath,
				};
			} else {
				this.logger.warn(`File "${f}" ignored, Venona loading only files that are mached to regexp ${runtimeRegexp.toString()}`);
				return Promise.resolve();
			}
		}));
		return _.compact(runtimes);
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
			const job = new Job(this.codefreshAPI, this.runtimes, taskLogger);
			this.logger.info(`Pushing job: ${Job.name} to queue`);
			this.queue.push(job, 1, this._handleJobError(job));
		});
	}

	_runOnce(Job) {
		if (!runOnce) {
			return;
		}
		const taskLogger = this.logger.child({
			namespace: LOGGER_NAMESPACES.TASK,
			job: Job.name,
			uid: new Chance().guid(),
		});
		const job = new Job(this.codefreshAPI, this.kubernetesAPI, taskLogger);
		job.exec();
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
