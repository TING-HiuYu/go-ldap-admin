<template>
  <div class="app-container">
    <div v-if="user">
      <Account :user="user" />
    </div>
  </div>
</template>

<script>
import { mapGetters } from 'vuex'
import Account from './components/Account'

export default {
  name: 'Profile',
  components: { Account },
  data() {
    return {
      user: {},
      activeTab: 'activity'
    }
  },
  computed: {
    ...mapGetters([
      'name',
      'avatar',
      'roles'
    ])
  },
  created() {
    this.getUser()
  },
  methods: {
    getUser() {
      this.user = {
        name: this.name,
        role: this.roles.join(' | '),
        email: 'admin@test.com',
        avatar: this.avatar
      }
    }
  }
}
</script>
