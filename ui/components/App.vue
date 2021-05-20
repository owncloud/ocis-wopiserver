<template>
  <div class="uk-flex uk-flex-column uk-flex-middle uk-height-1-1">

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
  mounted () {
    console.log('wopi mounted')
    this.loadDocument(this.filePath)
  },

  watch: {
    wopiClientUrl () {
      this.reloadWopi()
    }
  },

  computed: {
    ...mapState('Wopi', ['wopiClientUrl']),
    filePath () {
      return '/home' + this.$route.params.filePath
    }
  },
  methods: {
    ...mapActions('Wopi', ['loadDocument']),
    reloadWopi () {
      this.$nextTick(() => {
        document.getElementById('collabora-submit-form').submit()
      })
    }
  }
}
</script>
