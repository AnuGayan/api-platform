Feature: Deploying an LLM proxy from platform-api to the gateway
  As an API platform operator
  I want an LLM proxy created in platform-api to be served by the gateway data plane
  So that the AI control plane and data plane work together on every supported database.

  Background:
    Given the platform-api control plane and gateway data plane are running
    And I am authenticated to platform-api

  @llm
  Scenario: An LLM proxy deployed to a gateway proxies chat completions to the mock upstream
    Given an LLM provider backed by the mock OpenAI upstream
    And an LLM proxy for that provider
    When I deploy the LLM provider and proxy to the gateway
    Then the gateway serves chat completions for the proxy
    And the chat completion response is OpenAI-shaped
