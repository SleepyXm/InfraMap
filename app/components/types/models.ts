export type InferenceProvider = {
  provider: string;
}

export type Model = {
  id: string;
  author: string;
  downloads: number;
  numParameters: number;
  pipeline_tag: string;
	availableInferenceProviders: InferenceProvider[];
  authorData: {
    avatarUrl: string;
    fullname: string;
  };
}