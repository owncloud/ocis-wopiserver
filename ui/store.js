const state = {
  config: null
}

const getters = {
  config: state => state.config,
  getServerForJsClient: (state, getters, rootState, rootGetters) => rootGetters.configuration.server.replace(/\/$/, ''),
  accessToken: (state, getters, rootState, rootGetters) => rootGetters.user.token
}

const actions = {
  // Used by ocis-web.
  loadConfig ({ commit }, config) {
    commit('LOAD_CONFIG', config)
  }
}

const mutations = {
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
