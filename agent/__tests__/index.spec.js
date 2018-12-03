const _ = require('lodash');
const Agent = require('./../');
const Codefresh = require('./../../services/Codefresh');
const Kubernetes = require('./../../services/Kubernetes');
const Logger = require('./../../services/Logger');
const { Server } = require('./../../server');


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
	}
});

describe('Agent unit test', () => {
	describe('Constructing new Agent', () => {
		it('Should construct successfully', () => {
			Kubernetes.buildFromConfig = jest.fn();
			const agent = new Agent(buildTestConfig());
			expect(Object.keys(agent).sort()).toEqual([
				'kubernetesAPI',
				'codefreshAPI',
				'name',
				'logger',
				'tasks',
				'server',
			].sort());
		});

		it('Should throw an error when config constructing anonymous agent', () => {
			try {
				// eslint-disable-next-line no-new
				new Agent();
			} catch (err) {
				expect(err.message).toEqual('Cannot construct anonymous agent');
			}
		});

		it('Should create logger during construction', () => {
			Kubernetes.buildFromConfig = jest.fn();
			new Agent(buildTestConfig());
			expect(Logger.create).toHaveBeenCalled();
		});

		it('Should create logger during construction with specific keys', () => {
			Kubernetes.buildFromConfig = jest.fn();
			new Agent(buildTestConfig());
			const callsArguments = Logger.create.mock.calls[0];
			expect(Object.keys(callsArguments[0])).toEqual(['name', 'version', 'mode']);
			expect(Object.keys(callsArguments[1])).toEqual(['prettyPrint', 'level']);
		});

		it('Should call logger with message during construction', () => {
			Kubernetes.buildFromConfig = jest.fn();
			new Agent(buildTestConfig());
			const callsArguments = Logger.create.mock.instances[1].info.mock.calls[0][0];
			expect(callsArguments).toEqual('Starting agent');
		});


		it('Should Codefresh API service during construction just once', () => {
			Kubernetes.buildFromConfig = jest.fn();
			new Agent(buildTestConfig());
			const totalCallsToCodefreshConstructor = Codefresh.mock.calls;
			expect(totalCallsToCodefreshConstructor).toHaveLength(1);
		});

		it('Should construct CodefreshAPI service with specific keys', () => {
			Kubernetes.buildFromConfig = jest.fn();
			new Agent(buildTestConfig());
			const callsArguments = Codefresh.mock.calls[0];
			expect(callsArguments).toHaveLength(2);
			expect(Object.keys(callsArguments[0])).toEqual(['name', 'version', 'mode']);
			expect(Object.keys(callsArguments[1])).toEqual(['baseURL', 'token']);
		});

		it('Should construct KubernetesAPI service with specific keys', () => {
			Kubernetes.buildFromConfig = jest.fn();
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

	describe('Initializing agent', () => {
		it('Should complete initialization with message', () => {
			Kubernetes.buildFromConfig = jest.fn(Kubernetes);
			Server.mockImplementationOnce(() => ({
				init: jest.fn(),
			}));
			Codefresh.mockImplementationOnce(() => ({
				init: jest.fn(),
			}));
			Kubernetes.mockImplementationOnce(() => ({
				init: jest.fn(),
			}));
			return new Agent(buildTestConfig())
				.init()
				.then(() => {
					const loggerSuccessMessage = Logger.create.mock.instances[1].info.mock.calls[2][0];
					expect(loggerSuccessMessage).toEqual('Initializing finished');
				});
		});

		it('Should call to Server initialization process during agent initialization', () => {
			const serverInitSpy = jest.fn();
			Server.mockImplementationOnce(() => ({
				init: serverInitSpy,
			}));
			Codefresh.mockImplementationOnce(() => ({
				init: jest.fn(),
			}));
			Kubernetes.mockImplementationOnce(() => ({
				init: jest.fn(),
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
			Server.mockImplementation(() => ({
				init: jest.fn(),
			}));
			Codefresh.mockImplementation(() => ({
				init: codefreshInitSpy,
			}));
			Kubernetes.mockImplementation(() => ({
				init: jest.fn(),
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
			Server.mockImplementation(() => ({
				init: jest.fn(),
			}));
			Codefresh.mockImplementation(() => ({
				init: jest.fn(),
			}));
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

		it('Initialization should return with no value', () => {
			Server.mockImplementation(() => ({
				init: jest.fn(),
			}));
			Codefresh.mockImplementation(() => ({
				init: jest.fn(),
			}));
			Kubernetes.mockImplementation(() => ({
				init: jest.fn(),
			}));
			return new Agent(buildTestConfig()).init()
				.then((...params) => {
					expect(...params).toBeUndefined();
				});
		});

		it('Should throw an error when initialization crashed', () => {
			Server.mockImplementation(() => ({
				init: jest.fn(() => Promise.reject(new Error('Error!!'))),
			}));
			Codefresh.mockImplementation(() => ({
				init: jest.fn(),
			}));
			Kubernetes.mockImplementation(() => ({
				init: jest.fn(),
			}));
			return new Agent(buildTestConfig()).init()
				.catch((err) => {
					expect(err.message).toEqual(expect.stringContaining('Failed to initialize agent with error message:'));
				});
		});
	});
});
