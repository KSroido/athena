import axios from 'axios'

const api = axios.create({
  baseURL: '/api',
  timeout: 30000,
  headers: { 'Content-Type': 'application/json' },
})

// Chat with AgentServer
export function chat(message: string) {
  return api.post('/chat', { message })
}

// Projects
export function listProjects() {
  return api.get('/projects')
}

export function createProject(name: string, originalRequirement: string, description?: string) {
  return api.post('/projects', { name, original_requirement: originalRequirement, description })
}

export function getProject(id: string) {
  return api.get(`/projects/${id}`)
}

// Blackboard
export function getBlackboard(projectId: string, category?: string) {
  const params = category ? { category } : {}
  return api.get(`/projects/${projectId}/blackboard`, { params })
}

export function writeBlackboard(
  projectId: string,
  category: string,
  content: string,
  certainty: string = 'conjecture',
  author?: string,
) {
  return api.post(`/projects/${projectId}/blackboard`, {
    category,
    content,
    certainty,
    author,
  })
}

// Agents
export function listAgents() {
  return api.get('/agents')
}

export function listProjectAgents(projectId: string) {
  return api.get(`/projects/${projectId}/agents`)
}

export default api
