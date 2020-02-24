const kube = require('kubernetes-client');
const Kubernetes = require('../Kubernetes');
const { create: createLogger } = require('./../../services/Logger');

jest.mock('./../../services/Logger');
jest.mock('kubernetes-client');

const getFakeMetadata = () => ({ name: 'unit-test' });

const getFakeValidConfig = () => ({
	config: {
		url: 'fake',
		auth: {
			bearer: 'bearer-token',
		},
		ca: 'ca-cert',
	},
});

function buildKubernetesAPI() {
	return Kubernetes.buildFromConfig(getFakeMetadata(), getFakeValidConfig());
}

describe('Kubernetes API unit tests', () => {
	describe('Construction', () => {
		it('Shoud construct from config', () => {
			buildKubernetesAPI();
			const callToClientConstructor = kube.Client.prototype.constructor.mock.calls;
			const paramsToConstructor = callToClientConstructor[0][0];
			expect(callToClientConstructor).toHaveLength(1);
			expect(paramsToConstructor).toHaveProperty('config.auth.bearer');
			expect(paramsToConstructor).toHaveProperty('config.ca');
			expect(paramsToConstructor).toHaveProperty('config.url');
		});

		it('Should throw error when url is not given', () => expect(() => Kubernetes.buildFromConfig(getFakeMetadata(), { config: {} })).toThrow('Failed to construct Kubernetes API service, missing Kubernetes URL'));

		it('Should throw error when bearer token is not given', () => expect(() => Kubernetes.buildFromConfig(getFakeMetadata(), { config: { url: 'ok' } })).toThrow('Failed to construct Kubernetes API service, missing Kubernetes bearer token'));

		it('Should throw error when ca certificate is not given', () => expect(() => Kubernetes.buildFromConfig(getFakeMetadata(), { config: { url: 'ok', auth: { bearer: 'token' } } })).toThrow('Failed to construct Kubernetes API service, missing Kubernetes ca certificate'));

		it('Should set value on this', () => {
			const api = buildKubernetesAPI();
			expect(Object.keys(api).sort()).toEqual(['client', 'metadata'].sort());
		});
	});

	describe('Initialization', () => {
		it('Should finish initialization', () => {
			const spy = jest.fn();
			kube.Client.mockImplementationOnce(() => ({
				loadSpec: spy,
			}));
			const api = buildKubernetesAPI();
			return api.init()
				.then(() => {
					expect(spy).toHaveBeenCalled();
				});
		});

		it('Should throw an error', () => {
			const spy = jest.fn().mockRejectedValue(new Error('error!'));
			kube.Client.mockImplementationOnce(() => ({
				loadSpec: spy,
			}));
			const api = buildKubernetesAPI();
			return expect(api.init()).rejects.toThrowError(new RegExp('Failed to complete Kubernetes service initialization with error'));
		});
	});

	describe('createPod', () => {
		it('Should create a pod successfully', () => {
			const desiredNamespace = 'fake-namespace';
			const fakePod = {
				metadata: {
					namespace: desiredNamespace,
					name: 'pod',
				},
			};
			const createPodSpy = jest.fn();
			const getNamespaceSpy = jest.fn();
			kube.Client.mockImplementationOnce(() => ({
				api: {
					v1: {
						namespaces: getNamespaceSpy.mockImplementation(() => ({
							pod: {
								post: createPodSpy,
							},
						})),
					},
				},
			}));
			return buildKubernetesAPI()
				.createPod(createLogger(), fakePod)
				.then(() => {
					expect(getNamespaceSpy).toHaveBeenCalledWith(desiredNamespace);
					expect(createPodSpy).toHaveBeenCalledWith({ body: fakePod });
				});
		});

		it('Should throw an error', () => {
			kube.Client.mockImplementationOnce(() => ({
				api: {
					v1: {
						namespaces: jest.fn(() => ({
							pod: {
								post: jest.fn().mockRejectedValue(new Error('Error to make api call')),
							},
						})),
					},
				},
			}));
			return expect(buildKubernetesAPI().createPod(createLogger(), { metadata: { name: 'fake-name' } })).rejects.toThrowError('Failed to create Kubernetes pod with message');
		});
	});

	describe('deletePod', () => {
		it('Should delete a pod successfully', () => {
			const namespace = 'fake-namespace';
			const name = 'pod';
			const fakePod = {
				metadata: {
					namespace,
					name,
				},
			};
			const deletePodSpy = jest.fn();
			const getNamespaceSpy = jest.fn();
			const podSpy = jest.fn();
			kube.Client.mockImplementationOnce(() => ({
				api: {
					v1: {
						namespaces: getNamespaceSpy.mockImplementation(() => ({
							pod: podSpy.mockImplementationOnce(() => {
								return {
									delete: deletePodSpy,
								};
							})
						})),
					},
				},
			}));
			return buildKubernetesAPI()
				.deletePod(createLogger(), fakePod.metadata.namespace, fakePod.metadata.name)
				.then(() => {
					expect(getNamespaceSpy).toHaveBeenCalledWith(namespace);
					expect(podSpy).toHaveBeenCalledWith(name);
					expect(deletePodSpy).toHaveBeenCalledWith();
				});
		});

		it('Should throw an error when pod deletion failed', () => {
			kube.Client.mockImplementationOnce(() => ({
				api: {
					v1: {
						namespaces: jest.fn(() => ({
							pod: jest.fn(() => {
								return {
									delete: jest.fn().mockRejectedValue(new Error('Error to make api call')),
								};
							}),
						})),
					},
				},
			}));
			return expect(buildKubernetesAPI().deletePod(createLogger(), '', '')).rejects.toThrowError('Failed to delete Kubernetes pod with message');
		});
	});

	describe('createPvc', () => {
		it('Should create a pvc successfully', () => {
			const desiredNamespace = 'fake-namespace';
			const fakePvc = {
				metadata: {
					namespace: desiredNamespace,
					name: 'pvc',
				},
			};
			const createPvcSpy = jest.fn();
			const getNamespaceSpy = jest.fn();
			kube.Client.mockImplementationOnce(() => ({
				api: {
					v1: {
						namespaces: getNamespaceSpy.mockImplementation(() => ({
							persistentvolumeclaim: {
								post: createPvcSpy,
							},
						})),
					},
				},
			}));
			return buildKubernetesAPI()
				.createPvc(createLogger(), fakePvc)
				.then(() => {
					expect(getNamespaceSpy).toHaveBeenCalledWith(desiredNamespace);
					expect(createPvcSpy).toHaveBeenCalledWith({ body: fakePvc });
				});
		});

		it('Should throw an error', () => {
			kube.Client.mockImplementationOnce(() => ({
				api: {
					v1: {
						namespaces: jest.fn(() => ({
							persistentvolumeclaim: {
								post: jest.fn().mockRejectedValue(new Error('Error to make api call')),
							},
						})),
					},
				},
			}));
			return expect(buildKubernetesAPI().createPvc(createLogger(), { metadata: { name: 'fake-name' } })).rejects.toThrowError('Failed to create Kubernetes pvc with message');
		});
	});

	describe('deletePvc', () => {
		it('Should delete a pod successfully', () => {
			const namespace = 'fake-namespace';
			const name = 'pod';
			const fakePod = {
				metadata: {
					namespace,
					name,
				},
			};
			const deletePvcSpy = jest.fn();
			const getNamespaceSpy = jest.fn();
			const podSpy = jest.fn();
			kube.Client.mockImplementationOnce(() => ({
				api: {
					v1: {
						namespaces: getNamespaceSpy.mockImplementation(() => ({
							persistentvolumeclaim: podSpy.mockImplementationOnce(() => {
								return {
									delete: deletePvcSpy,
								};
							})
						})),
					},
				},
			}));
			return buildKubernetesAPI()
				.deletePvc(createLogger(), fakePod.metadata.namespace, fakePod.metadata.name)
				.then(() => {
					expect(getNamespaceSpy).toHaveBeenCalledWith(namespace);
					expect(podSpy).toHaveBeenCalledWith(name);
					expect(deletePvcSpy).toHaveBeenCalledWith();
				});
		});

		it('Should throw an error when pod deletion failed', () => {
			kube.Client.mockImplementationOnce(() => ({
				api: {
					v1: {
						namespaces: jest.fn(() => ({
							persistentvolumeclaim: jest.fn(() => {
								return {
									delete: jest.fn().mockRejectedValue(new Error('Error to make api call')),
								};
							}),
						})),
					},
				},
			}));
			return expect(buildKubernetesAPI().deletePvc(createLogger(), '', '')).rejects.toThrowError('Failed to delete Kubernetes pvc with message');
		});
	});
});
