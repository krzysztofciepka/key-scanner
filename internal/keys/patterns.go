package keys

type Pattern struct {
	EnvVar   string
	Provider string
}

var BuiltinPatterns = []Pattern{
	{"OPENAI_API_KEY", "OpenAI"},
	{"ANTHROPIC_API_KEY", "Anthropic"},
	{"OPENROUTER_API_KEY", "OpenRouter"},
	{"OPENCODE_API_KEY", "OpenCode"},
	{"GOOGLE_API_KEY", "Google AI"},
	{"GEMINI_API_KEY", "Google AI"},
	{"GROQ_API_KEY", "Groq"},
	{"TOGETHER_API_KEY", "Together AI"},
	{"DEEPSEEK_API_KEY", "DeepSeek"},
	{"COHERE_API_KEY", "Cohere"},
	{"REPLICATE_API_KEY", "Replicate"},
	{"HF_API_KEY", "HuggingFace"},
	{"HUGGINGFACE_API_KEY", "HuggingFace"},
	{"MISTRAL_API_KEY", "Mistral"},
	{"PERPLEXITY_API_KEY", "Perplexity"},
	{"BLACKBOX_AI_KEY", "Blackbox AI"},
	{"XAI_API_KEY", "xAI"},
	{"FIREWORKS_API_KEY", "Fireworks"},
	{"GITHUB_TOKEN", "GitHub"},
	{"GH_TOKEN", "GitHub"},
}

func FilterByEnvVar(envVar string) []Pattern {
	if envVar == "" {
		return BuiltinPatterns
	}
	for _, p := range BuiltinPatterns {
		if p.EnvVar == envVar {
			return []Pattern{p}
		}
	}
	return nil
}
