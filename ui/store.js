import axios from 'axios'

const state = {
  config: null,
  wopiClientUrl: ''
}

const getters = {
  config: state => state.config,
  getServerForJsClient: (state, getters, rootState, rootGetters) => rootGetters.configuration.server.replace(/\/$/, '')
}

const actions = {
  // Used by ocis-web.
  loadConfig ({ commit }, config) {
    commit('LOAD_CONFIG', config)
  },

  loadDocument ({ commit, dispatch, getters, rootGetters }, filePath) {
    injectAuthToken(rootGetters)

    const tokenUrl = getters.getServerForJsClient + '/api/v0/wopi/open'

    axios.get(tokenUrl, { params: { filePath: filePath } })
      .then(response => {
        commit('SET_DOCUMENT', { wopiClientUrl: response.data.wopiclienturl })
      })
      .catch(error => {
        this.errorMessage = error.message
        console.error('There was an error!', error)
      })
  }
}

const mutations = {
  SET_DOCUMENT (state, { wopiClientUrl }) {
    state.wopiClientUrl = wopiClientUrl
  },
  LOAD_CONFIG (state, config) {
    state.config = config
  }
}

export default {
  namespaced: true,
  state,
  getters,
  actions,
  mutations
}

function injectAuthToken (rootGetters) {
  axios.interceptors.request.use(config => {
    if (typeof config.headers.Authorization === 'undefined') {
      const token = rootGetters.user.token
      if (token) {
        config.headers.Authorization = `Bearer ${token}`
      }
    }
    return config
  })
}
