import 'regenerator-runtime/runtime'
import axios from 'axios'

import store from './store'

// file extensions that are supported to be opened
const fileExtensions = [
  'odt',
  'ott',
  'ods',
  'odp',
  'odg',
  'otg',
  'doc',
  'dot',
  'xls',
  'xlt',
  'xlm',
  'ppt',
  'pot',
  'pps',
  'vsd',
  'dxf',
  'wmf',
  'cdr',
  'pages',
  'number',
  'key'
]

// file extensions that are working to create new files
const newFileExtensions = [
  'odt',
  'ods',
  'odp',
  'odg'
]

const openModes = [
  'edit'
]

const appInfo = {
  name: 'Wopi',
  id: 'wopi',
  isFileEditor: true,
  icon: 'x-office-document',
  extensions: getExtensions(openModes, fileExtensions)
}

export default {
  appInfo,
  store
}

function getExtension (openMode, fileExtension) {
  let newFileMenu = null
  if (newFileExtensions.includes(fileExtension)) {
    newFileMenu = {
      menuTitle ($gettext) {
        return $gettext('New ' + fileExtension.toUpperCase() + ' document')
      }
    }
  }
  return {
    extension: fileExtension,
    icon: 'x-office-document',
    routes: [
      'files-personal',
      'files-favorites',
      'files-shared-with-others',
      'files-shared-with-me'
    ],
    handler: function ({ extensionConfig, filePath, fileId }) {
      axios.interceptors.request.use(config => {
        if (typeof config.headers.Authorization === 'undefined') {
          if (window.Vue.$store.getters['Wopi/accessToken']) {
            config.headers.Authorization = 'Bearer ' + window.Vue.$store.getters['Wopi/accessToken']
          }
        }
        return config
      })
      const tokenUrl = window.Vue.$store.getters['Wopi/getServerForJsClient'] + '/api/v0/wopi/open'
      axios.get(tokenUrl, { params: { filePath: '/home' + filePath, fileId: fileId } })
        .then(response => {
          var form = document.createElement('form')
          form.setAttribute('method', 'POST')
          form.setAttribute('action', response.data.wopiclienturl)
          form.setAttribute('target', '_blank')

          var accesstoken = document.createElement('input')
          accesstoken.type = 'hidden'
          accesstoken.name = 'access_token'
          accesstoken.value = response.data.accesstoken
          form.appendChild(accesstoken)

          var accesstokenttl = document.createElement('input')
          accesstokenttl.type = 'hidden'
          accesstokenttl.name = 'access_token_ttl'
          accesstokenttl.value = response.data.accesstokenttl
          form.appendChild(accesstokenttl)

          var f = document.body.appendChild(form)
          form.submit()
          document.body.removeChild(f)
        })
        .catch(error => {
          this.errorMessage = error.message
          console.error('opening file with WOPI failed', error)
        })
    },
    newFileMenu: newFileMenu
  }
}

function getExtensions (openModes, fileExtensions) {
  const ext = []
  openModes.forEach(
    function (m) {
      fileExtensions.forEach(
        function (e) {
          ext.push(getExtension(m, e)
          )
        }
      )
    }
  )
  return ext
}
