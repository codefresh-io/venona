
const Promise = require('bluebird');
const scheduler = require('node-schedule');
const fs = require('fs');
const path = require('path');
const _ = require('lodash');
const recursive = require('recursive-readdir');
const Agent = require('./../');
const Codefresh = require('./../../services/Codefresh');
const Logger = require('./../../services/Logger');
const { Server } = require('./../../server');
const TaskPullerJob = require('./../../jobs/TaskPullerJob/TaskPuller.job');
const StatusReporterJob = require('./../../jobs/StatusReporterJob/StatusReporter.job');
const Kubernetes = require('./../../services/Kubernetes');

jest.mock('./../../services/Codefresh');
jest.mock('./../../services/Logger');
jest.mock('./../../services/Kubernetes');
jest.mock('./../../server');
jest.mock('fs');
jest.mock('recursive-readdir');

const buildTestConfig = () => ({
	metadata: {
		name: 'agent',
		version: '1.0',
		mode: 'mode',
		venonaConfDir: '/path/to/venona/config/dir'
	},
	server: {
		port: '9000',
	},
	logger: {
		prettyPrint: true,
		level: 'info',
	},
	kubernetes: {
		config: {
			url: 'host',
			auth: {
				bearer: 'token'
			},
			ca: 'ca'
		},
	},
	codefresh: {
		baseURL: 'https://g.codefresh.io',
		token: 'token'
	},
	jobs: {
		TaskPullerJob: {
			cronExpression: 'cron'
		},
		StatusReporterJob: {
			cronExpression: 'cron',
		},
		queue: {

		}
	}
});

const loadActualJobs = () => {
	recursive.__setFiles([
		path.join(__dirname, '../../', 'jobs/StatusReporterJob/StatusReporter.job.js'),
		path.join(__dirname, '../../', 'jobs/TaskPullerJob/TaskPuller.job.js'),
	]);
};

beforeEach(() => {
	Server.mockImplementationOnce(() => ({
		init: jest.fn(),
	}));
	Codefresh.mockImplementationOnce(() => ({
		init: jest.fn(),
	}));
	recursive.__setFiles([]);
	fs.readdir.mockImplementationOnce((path, cb) => {
		cb(null, []);
	});
});

afterEach(() => {
	recursive.__clear();
});

describe('Agent unit test', () => {

	describe('Initializing agent', () => {

		describe('positive', () => {
			it('Should report all services been initialized', async () => {
				const agent = new Agent();
				await agent.init(buildTestConfig());
				const loggerSuccessMessage = Logger.create.mock.instances[1].info.mock.calls[2][0];
				expect(loggerSuccessMessage).toEqual('All services has been initialized');

			});

			it('Should call to Server initialization process during agent initialization', async () => {
				const serverInitSpy = jest.fn();
				Server.mockReset();
				Server.mockImplementationOnce(() => ({
					init: serverInitSpy,
				}));
				const agent = new Agent();
				await agent.init(buildTestConfig());
				expect(serverInitSpy).toHaveBeenCalledTimes(1);
				expect(serverInitSpy).toHaveBeenCalledWith();

			});

			it('Should call to CodefreshAPI initialization process during agent initialization', async () => {
				const codefreshInitSpy = jest.fn();
				Codefresh.mockReset();
				Codefresh.mockImplementation(() => ({
					init: codefreshInitSpy,
					reportStatus: jest.fn(),
				}));
				jest.unmock('recursive-readdir');
				const agent = new Agent();
				await agent.init(buildTestConfig());
				expect(codefreshInitSpy).toHaveBeenCalledTimes(1);
				expect(codefreshInitSpy).toHaveBeenCalledWith();

			});

			it('Should call to _startJob for both supported tasks', async () => {
				loadActualJobs();
				const agent = new Agent();
				agent._startJob = jest.fn();
				await agent.init(buildTestConfig());
				expect(agent._startJob).toHaveBeenCalledTimes(2);
				expect(agent._startJob).toHaveBeenNthCalledWith(1, StatusReporterJob);
				expect(agent._startJob).toHaveBeenNthCalledWith(2, TaskPullerJob);

			});
			it('Should call task that has runOnce when agent starts ', async () => {
				loadActualJobs();
				const config = buildTestConfig();
				config.jobs.StatusReporterJob.runOnce = true;
				const agent = new Agent();
				agent._runOnce = jest.fn();
				agent._startJob = jest.fn();
				await agent.init(config);
				expect(agent._runOnce).toHaveBeenCalledTimes(1);
				expect(agent._runOnce).toHaveBeenNthCalledWith(1, StatusReporterJob);

			});
		});
		
		describe('Prepare server', () => {
			it('Should throw an error when initialization crashed', () => {
				Server.mockReset();
				Server.mockImplementationOnce(() => ({
					init: jest.fn().mockRejectedValue(new Error('Error!')),
				}));
				return expect(new Agent().init(buildTestConfig())).rejects.toThrow('Failed to initialize agent with error, message: Error!');
			});
		});

		describe('Prepare runtimes', () => {
			it('Should fail to init in case failed to read metadata.venonaConfDir directory', async () => {
				fs.readdir.mockReset();
				fs.readdir.mockImplementationOnce((path, cb) => cb(new Error('Failed to read directory')));
				const agent = new Agent();
				return expect(agent.init(buildTestConfig())).rejects.toThrow('Failed to initialize agent with error, message: Failed to read directory');
			});
			it('Should log warning that nasted directory under metadata.venonaConfDir is not going to be read', async () => {
				fs.readdir.mockReset();
				fs.readdir.mockImplementationOnce((path, cb) => cb(null, [
					'sub-dir',
				]));
				fs.stat.mockImplementation((path, cb) => {
					if (path.includes('sub-di')) {
						cb(null, {
							isDirectory: () => true,
						});
					}
				});

				const agent = new Agent();
				await agent.init(buildTestConfig());
				expect(agent.logger.warn).toHaveBeenCalledTimes(1);
				expect(agent.logger.warn).toHaveBeenCalledWith('Directory "sub-dir" ignored, Venona loading only files that are mached to regexp /.*\\.runtime\\.yaml/');
			});
			it('Should log warning that file that is not matched to the regexp is ignored from being loaded', async () => {
				fs.readdir.mockReset();
				fs.readdir.mockImplementationOnce((path, cb) => cb(null, [
					'file',
				]));
				fs.stat.mockImplementation((path, cb) => {
					cb(null, {
						isDirectory: () => false,
					});
				});

				const agent = new Agent();
				await agent.init(buildTestConfig());
				expect(agent.logger.warn).toHaveBeenCalledTimes(1);
				expect(agent.logger.warn).toHaveBeenCalledWith('File "file" ignored, Venona loading only files that are mached to regexp /.*\\.runtime\\.yaml/');
			});
			it('Should log warning in case found duplication of runtimes', async () => {
				const db = {
					'a.runtime.yaml': 'name: runtime',
					'b.runtime.yaml': 'name: runtime',
				};
				fs.readdir.mockReset();
				fs.readdir.mockImplementationOnce((path, cb) => {
					cb(null, _.keys(db));
				});
				fs.stat.mockImplementation((path, cb) => {
					cb(null, {
						isDirectory: () => false,
					});
				});
				fs.readFile.mockImplementation((location, cb) => {
					cb(null, db[path.basename(location)]);
				});
				fs.access.mockImplementation((path, cb) => {
					cb(null);
				});
				Kubernetes.mockImplementation(() => ({
					init: Promise.resolve,
				}));
				Kubernetes.buildFromConfig = jest.fn(Kubernetes);

				const agent = new Agent();
				await expect(agent.init(buildTestConfig())).resolves.toEqual();
				expect(agent.logger.warn).toHaveBeenCalledWith('Ignoring runtime "runtime", already been configured, conflict with file: "/path/to/venona/config/dir/b.runtime.yaml"');
			});
		});
	});

	describe('Queue', () => {

		it('Should call run on each job in the queue', async () => {
			scheduler.scheduleJob = jest.fn((_ex, cb) => {
				cb();
			});
			const spy = jest.fn();
			const agent = new Agent();
			await agent.init(buildTestConfig());
			const FakeJob = function() {
				this.exec = spy;
				return this;
			};
			agent._startJob(FakeJob);
			await Promise.delay(2000);
			expect(spy).toHaveBeenCalledTimes(1);
			expect(spy).toHaveBeenCalledWith();
		});

		it('Should call cb with an error in case a job been rejected', async () => {
			scheduler.scheduleJob = jest.fn((_ex, cb) => {
				cb();
			});
			const error = new Error('my error');
			const agent = new Agent();
			await agent.init(buildTestConfig());
			const FakeJob = function() {
				this.exec = jest.fn().mockRejectedValue(error);
				return this;
			};
			const handleErrorSpy = jest.fn();
			agent._handleJobError = jest.fn(() => {
				return handleErrorSpy;
			});
			agent._startJob(FakeJob);
			await Promise.delay(100);
			expect(handleErrorSpy).toHaveBeenCalledTimes(1);
			expect(handleErrorSpy).toHaveBeenCalledWith(error);
		});

		it('Should call cb with an error in case a job throws an error', async () => {
			scheduler.scheduleJob = jest.fn((_ex, cb) => {
				cb();
			});
			const error = new Error('my error');
			const agent = new Agent();
			await agent.init(buildTestConfig());
			const FakeJob = function() {
				this.exec = jest.fn(() => {
					throw (error);
				});
				return this;
			};
			const handleErrorSpy = jest.fn();
			agent._handleJobError = jest.fn(() => {
				return handleErrorSpy;
			});
			agent._startJob(FakeJob);
			await Promise.delay(100);
			expect(handleErrorSpy).toHaveBeenCalledTimes(1);
			expect(handleErrorSpy).toHaveBeenCalledWith(error);
		});

		it('Should call cb in case the job was resolved', async () => {
			scheduler.scheduleJob = jest.fn((_ex, cb) => {
				cb();
			});
			const agent = new Agent();
			await agent.init(buildTestConfig());
			const FakeJob = function() {
				this.exec = jest.fn();
				return this;
			};
			const handleErrorSpy = jest.fn();
			agent._handleJobError = jest.fn(() => {
				return handleErrorSpy;
			});
			agent._startJob(FakeJob);
			await Promise.delay(100);
			expect(handleErrorSpy).toHaveBeenCalledTimes(1);
			expect(handleErrorSpy).toHaveBeenCalledWith();
		});

		it('Should call to _onEmptyQueue once a queue have no more tasks', async () => {
			scheduler.scheduleJob = jest.fn((_ex, cb) => {
				cb();
			});
			const agent = new Agent();
			await agent.init(buildTestConfig());
			const FakeJob = function() {
				this.exec = jest.fn();
				return this;
			};
			const handleErrorSpy = jest.fn();
			agent._handleJobError = jest.fn(() => {
				return handleErrorSpy;
			});
			agent._startJob(FakeJob);
			await Promise.delay(100);
			expect(agent.logger.info).toHaveBeenCalledWith('Queue is empty');
		});
	});

	describe('Auto jobs load', () => {
		it('Should load only jobs that related to agent', async () => {
			loadActualJobs();
			const agent = new Agent();
			agent._startJob = jest.fn();
			await agent.init(buildTestConfig());
			expect(agent._startJob).toHaveBeenCalledTimes(2);
		});
	});

	describe('_startJob', () => {

		it('Should create fork logger for each job', async () => {
			scheduler.scheduleJob = jest.fn((_ex, cb) => {
				cb();
			});
			const agent = new Agent();
			await agent.init(buildTestConfig());
			const FakeJob = function() {
				this.exec = jest.fn();
				return this;
			};
			agent.logger.child.mockReset();
			agent._startJob(FakeJob);
			await Promise.delay(100);
			expect(agent.logger.child).toHaveBeenCalledTimes(1);
		});

		it('Should pass all services and runtimes the Job constructor', async () => {
			scheduler.scheduleJob = jest.fn((_ex, cb) => {
				cb();
			});
			const agent = new Agent();
			await agent.init(buildTestConfig());
			const FakeJob = jest.fn().mockImplementation(() => {
				this.exec = jest.fn();
				return this;
			});
			const newTaskLogger = {};
			agent.logger.child.mockImplementationOnce(() => {
				return newTaskLogger;
			});
			agent._startJob(FakeJob);
			await Promise.delay(100);
			expect(FakeJob).toHaveBeenCalledWith(agent.codefreshAPI, {}, newTaskLogger);
		});

		it('Should push the job to the queue', async () => {
			scheduler.scheduleJob = jest.fn((_ex, cb) => {
				cb();
			});
			const agent = new Agent();
			await agent.init(buildTestConfig());
			const FakeJob = function() {
				this.exec = jest.fn();
				return this;
			};
			agent.queue.push = jest.fn();
			agent._startJob(FakeJob);
			await Promise.delay(100);
			expect(agent.queue.push).toHaveBeenCalledTimes(1);
		});
	});
});
