module Gitlab
  module Git
    class LfsChanges
      def initialize(repository, newrev)
        @repository = repository
        @newrev = newrev
      end

      def new_pointers(object_limit: nil, not_in: nil)
        git_new_pointers(object_limit, not_in)
      end

      def all_pointers
        git_all_pointers
      end

      private

      def git_new_pointers(object_limit, not_in)
        rev_list.new_objects(rev_list_params(not_in: not_in)) do |object_ids|
          object_ids = object_ids.take(object_limit) if object_limit

          Gitlab::Git::Blob.batch_lfs_pointers(@repository, object_ids)
        end
      end

      def git_all_pointers
        params = {}
        params[:options] = ["--filter=blob:limit=#{Gitlab::Git::Blob::LFS_POINTER_MAX_SIZE}"] if rev_list_supports_new_options?

        rev_list.all_objects(rev_list_params(params)) do |object_ids|
          Gitlab::Git::Blob.batch_lfs_pointers(@repository, object_ids)
        end
      end

      def rev_list
        Gitlab::Git::RevList.new(@repository, newrev: @newrev)
      end

      # We're passing the `--in-commit-order` arg to ensure we don't wait
      # for git to traverse all commits before returning pointers.
      # This is required in order to improve the performance of LFS integrity check
      def rev_list_params(params = {})
        params[:options] ||= []
        params[:options] << "--in-commit-order" if rev_list_supports_new_options?
        params[:require_path] = true

        params
      end

      def rev_list_supports_new_options?
        return @option_supported if defined?(@option_supported)

        @option_supported = Gitlab::Git.version >= Gitlab::VersionInfo.parse('2.16.0')
      end
    end
  end
end
