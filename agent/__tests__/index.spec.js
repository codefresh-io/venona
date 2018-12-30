const _ = require('lodash');
const Promise = require('bluebird');
const scheduler = require('node-schedule');
const Agent = require('./../');
const Codefresh = require('./../../services/Codefresh');
const Kubernetes = require('./../../services/Kubernetes');
const Logger = require('./../../services/Logger');
const { Server } = require('./../../server');
const TaskPullerJob = require('./../../jobs/TaskPullerJob/TaskPuller.job');
const StatusReporterJob = require('./../../jobs/StatusReporterJob/StatusReporter.job');

jest.mock('./../../services/Codefresh');
jest.mock('./../../services/Kubernetes');
jest.mock('./../../services/Logger');
jest.mock('./../../server');

const buildTestConfig = () => ({
	metadata: {
		name: 'agent',
		version: '1.0',
		mode: 'mode',
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
			cronExpression: 'cron'
		},
		queue: {

		}
	}
});

beforeEach(() => {
	Server.mockImplementationOnce(() => ({
		init: jest.fn(),
	}));
	Codefresh.mockImplementationOnce(() => ({
		init: jest.fn(),
	}));
	Kubernetes.mockImplementationOnce(() => ({
		init: jest.fn(),
	}));
	Kubernetes.buildFromConfig = jest.fn(Kubernetes);
	Kubernetes.buildFromInCluster = jest.fn(Kubernetes);
});

describe('Agent unit test', () => {
	describe('Constructing new Agent', () => {

		describe('positive', () => {
			it('Should construct successfully', () => {
				const agent = new Agent(buildTestConfig());
				expect(Object.keys(agent).sort()).toEqual([
					'kubernetesAPI',
					'codefreshAPI',
					'logger',
					'jobs',
					'queue',
					'server',
				].sort());
			});

			it('Should create logger during construction', () => {
				new Agent(buildTestConfig());
				expect(Logger.create).toHaveBeenCalled();
			});

			it('Should create logger during construction with specific keys', () => {
				new Agent(buildTestConfig());
				const callsArguments = Logger.create.mock.calls[0];
				expect(Object.keys(callsArguments[0])).toEqual(['name', 'version', 'mode']);
				expect(Object.keys(callsArguments[1])).toEqual(['prettyPrint', 'level']);
			});

			it('Should call logger with message during construction', () => {
				new Agent(buildTestConfig());
				const callsArguments = Logger.create.mock.instances[1].info.mock.calls[0][0];
				expect(callsArguments).toEqual('Starting agent');
			});

			it('Should Codefresh API service during construction just once', () => {
				new Agent(buildTestConfig());
				const totalCallsToCodefreshConstructor = Codefresh.mock.calls;
				expect(totalCallsToCodefreshConstructor).toHaveLength(1);
			});

			it('Should construct CodefreshAPI service with specific keys', () => {
				new Agent(buildTestConfig());
				const callsArguments = Codefresh.mock.calls[0];
				expect(callsArguments).toHaveLength(2);
				expect(Object.keys(callsArguments[0])).toEqual(['name', 'version', 'mode']);
				expect(Object.keys(callsArguments[1])).toEqual(['baseURL', 'token']);
			});

			it('Should construct KubernetesAPI service with specific keys', () => {
				new Agent(buildTestConfig());
				const callsArguments = Kubernetes.buildFromConfig.mock.calls[0];
				expect(callsArguments).toHaveLength(2);
				expect(Object.keys(callsArguments[0])).toEqual(['name', 'version', 'mode']);
				expect(Object.keys(callsArguments[1])).toEqual(['config']);
			});

			it('Should construct KubernetesAPI service when agent running from inside a cluster', () => {
				Kubernetes.buildFromInCluster = jest.fn();
				new Agent(_.merge(buildTestConfig(), { metadata: { name: 'fake-name', mode: 'InCluster' } }));
				const callsArguments = Kubernetes.buildFromInCluster.mock.calls[0];
				expect(Kubernetes.buildFromInCluster).toHaveBeenCalled();
				expect(Object.keys(callsArguments[0])).toEqual(['name', 'version', 'mode']);
			});

			it('Should construct Server with specific keys', () => {
				new Agent(buildTestConfig());
				const callsArguments = Server.mock.calls[0];
				expect(callsArguments).toHaveLength(3);
				expect(Object.keys(callsArguments[0])).toEqual(['name', 'version', 'mode']);
				expect(Object.keys(callsArguments[1])).toEqual(['port']);
				expect(Object.keys(callsArguments[2])).toEqual(['info', 'child', 'error']);
			});

		});

		describe('negative', () => {
			it('Should throw an error in case the agent was not constructed correctly', () => {
				try {
					Logger.create.mockImplementationOnce(() => {
						throw new Error('error');
					});
					new Agent(buildTestConfig());
				} catch (err) {
					expect(err.toString()).toEqual('Error: error');
				}
			});
		});

	});

	describe('Initializing agent', () => {

		describe('positive', () => {
			it('Should report all services been initialized', () => {
				return new Agent(buildTestConfig())
					.init()
					.then(() => {
						const loggerSuccessMessage = Logger.create.mock.instances[1].info.mock.calls[2][0];
						expect(loggerSuccessMessage).toEqual('All services has been initialized');
					});
			});

			it('Should call to Server initialization process during agent initialization', () => {
				const serverInitSpy = jest.fn();
				Server.mockReset();
				Server.mockImplementationOnce(() => ({
					init: serverInitSpy,
				}));
				return new Agent(buildTestConfig())
					.init()
					.then(() => {
						expect(serverInitSpy).toHaveBeenCalledTimes(1);
						expect(serverInitSpy).toHaveBeenCalledWith();
					});
			});

			it('Should call to CodefreshAPI initialization process during agent initialization', () => {
				const codefreshInitSpy = jest.fn();
				Codefresh.mockReset();
				Codefresh.mockImplementation(() => ({
					init: codefreshInitSpy,
				}));
				return new Agent(buildTestConfig())
					.init()
					.then(() => {
						expect(codefreshInitSpy).toHaveBeenCalledTimes(1);
						expect(codefreshInitSpy).toHaveBeenCalledWith();
					});
			});

			it('Should call to KubernetesAPI initialization process during agent initialization', () => {
				const kubernetesInitSpy = jest.fn();
				Kubernetes.mockReset();
				Kubernetes.mockImplementation(() => ({
					init: kubernetesInitSpy,
				}));
				return new Agent(buildTestConfig())
					.init()
					.then(() => {
						expect(kubernetesInitSpy).toHaveBeenCalledTimes(1);
						expect(kubernetesInitSpy).toHaveBeenCalledWith();
					});
			});

			it('Should call to _startJob for both supported tasks', () => {
				const agent = new Agent(buildTestConfig());
				agent._startJob = jest.fn();
				return agent
					.init()
					.then(() => {
						expect(agent._startJob).toHaveBeenCalledTimes(2);
						expect(agent._startJob).toHaveBeenNthCalledWith(1, StatusReporterJob);
						expect(agent._startJob).toHaveBeenNthCalledWith(2, TaskPullerJob);
					});
			});
		});

		describe('negative', () => {

			it('Should throw an error when initialization crashed', () => {
				Server.mockReset();
				Server.mockImplementationOnce(() => ({
					init: jest.fn().mockRejectedValue(new Error('Error!')),
				}));
				return expect(new Agent(buildTestConfig()).init()).rejects.toThrow('Failed to initialize agent with error message');
			});
		});

	});

	describe('Queue', () => {

		it('Should call run on each job in the queue', async () => {
			scheduler.scheduleJob = jest.fn((_ex, cb) => {
				cb();
			});
			const spy = jest.fn();
			const agent = new Agent(buildTestConfig());
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
			const agent = new Agent(buildTestConfig());
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
			const agent = new Agent(buildTestConfig());
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

		it.skip('Should call cb with an error in case a job timedout', () => {});

		it('Should call cb in case the job was resolved', async () => {
			scheduler.scheduleJob = jest.fn((_ex, cb) => {
				cb();
			});
			const agent = new Agent(buildTestConfig());
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
			const agent = new Agent(buildTestConfig());
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
			const agent = new Agent(buildTestConfig());
			agent._startJob = jest.fn();
			await agent._loadJobs();
			expect(agent._startJob).toHaveBeenCalledTimes(2);
		});
	});

	describe('_startJob', () => {

		it('Should create fork logger for each job', async () => {
			scheduler.scheduleJob = jest.fn((_ex, cb) => {
				cb();
			});
			const agent = new Agent(buildTestConfig());
			const FakeJob = function() {
				this.exec = jest.fn();
				return this;
			};
			agent.logger.child.mockReset();
			agent._startJob(FakeJob);
			await Promise.delay(100);
			expect(agent.logger.child).toHaveBeenCalledTimes(1);
		});

		it('Should pass all services to the Job constructor', async () => {
			scheduler.scheduleJob = jest.fn((_ex, cb) => {
				cb();
			});
			const agent = new Agent(buildTestConfig());
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
			expect(FakeJob).toHaveBeenCalledWith(agent.codefreshAPI, agent.kubernetesAPI, newTaskLogger);
		});

		it('Should push the job to the queue', async () => {
			scheduler.scheduleJob = jest.fn((_ex, cb) => {
				cb();
			});
			const agent = new Agent(buildTestConfig());
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
