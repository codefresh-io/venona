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

		it('Should crash the process in case the agent was not constructed correctly', () => {
			expect(true).toBeFalsy();
		});
	});

	describe('Initializing agent', () => {
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

		it('Should throw an error when initialization crashed', () => {
			Server.mockReset();
			Server.mockImplementationOnce(() => ({
				init: jest.fn().mockRejectedValue(new Error('Error!')),
			}));
			return expect(new Agent(buildTestConfig()).init()).rejects.toThrow('Failed to initialize agent with error message');
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

		it('Should crash the process in case the agent wasnt initialized correctly', () => {
			expect(true).toBeFalsy();
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
				this.fake = jest.fn();
				return this;
			};
			agent._startJob(FakeJob);
			await Promise.delay(2000);
			expect(spy).toHaveBeenCalledTimes(1);
			expect(spy).toHaveBeenCalledWith();
		});

		it('Should call cb with an error in case a job been rejected', () => {
			expect(true).toBeFalsy();
		});

		it('Should call cb with an error in case a job throws an error', () => {
			expect(true).toBeFalsy();
		});

		it('Should call cb with an error in case a job timedout', () => {
			expect(true).toBeFalsy();
		});

		it('Should call cb in case the job was resolved', () => {
			expect(true).toBeFalsy();
		});

		it('Should call to _onEmptyQueue once a queue have no more tasks', () => {
			expect(true).toBeFalsy();
		});
	});

	describe('Auto jobs load', () => {
		it('Should load only jobs that related to agent', () => {
			expect(true).toBeFalsy();
		});

		it('Should not load jobs that dosent matched pattern', () => {
			expect(true).toBeFalsy();
		});
	});

	describe('_startJob', () => {
		it('Should create fork logger for each job', () => {
			expect(true).toBeFalsy();
		});

		it('Should pass all services to the Job constructor', () => {
			expect(true).toBeFalsy();
		});

		it('Should push the job to the queue', () => {
			expect(true).toBeFalsy();
		});
	});
});
