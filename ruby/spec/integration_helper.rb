require 'socket'

require 'gitaly'
require 'spec_helper'

SOCKET_PATH = 'gitaly.socket'.freeze
GITALY_RUBY_DIR = File.expand_path('..', __dir__)
TMP_DIR = File.expand_path('../tmp', __dir__)

module IntegrationClient
  def gitaly_stub(service)
    klass = Gitaly.const_get(service).const_get(:Stub)
    klass.new("unix:tmp/#{SOCKET_PATH}", :this_channel_is_insecure)
  end

  def gitaly_repo(storage, relative_path)
    Gitaly::Repository.new(storage_name: storage, relative_path: relative_path)
  end
end

def start_gitaly
  build_dir = File.expand_path('../../_build', __dir__)
  gitlab_shell_dir = File.join(TMP_DIR, 'gitlab-shell')

  FileUtils.mkdir_p([TMP_DIR, File.join(gitlab_shell_dir, 'hooks')])

  config_toml = <<~CONFIG
    socket_path = "#{SOCKET_PATH}"
    bin_dir = "#{build_dir}/bin"

    [gitlab-shell]
    dir = "#{gitlab_shell_dir}"

    [gitaly-ruby]
    dir = "#{GITALY_RUBY_DIR}"

    [[storage]]
    name = "#{DEFAULT_STORAGE_NAME}"
    path = "#{DEFAULT_STORAGE_DIR}"
  CONFIG
  config_path = File.join(TMP_DIR, 'gitaly-rspec-config.toml')
  File.write(config_path, config_toml)

  test_log = File.join(TMP_DIR, 'gitaly-rspec-test.log')
  options = { out: test_log, err: test_log, chdir: TMP_DIR }

  gitaly_pid = spawn(File.join(build_dir, 'bin/gitaly'), config_path, options)
  at_exit { Process.kill('KILL', gitaly_pid) }

  wait_ready!(File.join('tmp', SOCKET_PATH))
end

def wait_ready!(socket)
  last_exception = StandardError.new('wait_ready! has not made any connection attempts')

  print('Booting gitaly for integration tests')
  100.times do |_i|
    sleep 0.1
    printf('.')
    begin
      UNIXSocket.new(socket).close
      puts ' ok'
      return
    rescue => ex
      last_exception = ex
    end
  end

  puts
  raise last_exception
end

start_gitaly
