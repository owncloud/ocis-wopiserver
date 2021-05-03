<template>
  <div class="uk-flex uk-flex-column uk-flex-middle uk-height-1-1">
    <h1>WOPI</h1>

    <form v-on:submit.prevent="openFile(filePathBox)" action="#" method="post">
      <oc-text-input v-model="filePathBox" placeholder="/home/Hello.odt" />

      <oc-button variation="primary" class="uk-width-1-1 uk-margin-top">
        Open
      </oc-button>
    </form>

    <div style="display: none">
      <!-- if you want to load it to the iframe below use this target: -->
      <!-- target="collabora-online-viewer" -->
      <form
        :action="wopiClientUrl"
        enctype="multipart/form-data"
        method="post"
        target="_blank"
        id="collabora-submit-form"
      >
        <input
          name="access_token"
          :value="accessToken"
          type="hidden"
          id="access-token"
        />
        <input type="submit" value="" />
      </form>
    </div>
    <iframe
      id="collabora-online-viewer"
      name="collabora-online-viewer"
      style="width: 90%; height: 80%; position: relative"
    >
    </iframe>
  </div>
</template>

<script>
import { mapActions, mapState } from 'vuex'

export default {
  name: 'App',
  data: function () {
    return {
      filePathBox: ''
    }
  },
  created () {
    // this.loadDocument(this.filePath)
  },

  watch: {
    wopiClientUrl () {
      this.reloadWopi()
    }
  },

  computed: {
    ...mapState('Wopi', ['wopiClientUrl', 'accessToken']),
    filePath () {
      return this.$route.params.filePath
    }
  },
  methods: {
    ...mapActions('Wopi', ['loadDocument']),
    reloadWopi () {
      this.$nextTick(() => {
        document.getElementById('collabora-submit-form').submit()
      })
    },
    openFile (targetFile) {
      this.$store.dispatch('Wopi/loadDocument', targetFile)
    }
  }
}
</script>
