FROM fluent/fluentd:v1.10-debian-1

# Use root account to use apt
USER root

WORKDIR /home/fluent
ENV PATH /fluentd/vendor/bundle/ruby/2.6.0/bin:$PATH
ENV GEM_PATH /fluentd/vendor/bundle/ruby/2.6.0
ENV GEM_HOME /fluentd/vendor/bundle/ruby/2.6.0
# skip runtime bundler installation
ENV FLUENTD_DISABLE_BUNDLER_INJECTION 1

# Install gem dependencies
COPY Gemfile* /fluentd/
RUN buildDeps="sudo make gcc g++ libc-dev" \
   && apt-get update \
   && apt-get install -y --no-install-recommends $buildDeps \
   && gem install bundler --version 2.1.4 \
   && bundle config silence_root_warning true \
   && bundle install --gemfile=/fluentd/Gemfile --path=/fluentd/vendor/bundle \
   && sudo gem sources --clear-all \
   && SUDO_FORCE_REMOVE=yes \
   apt-get purge -y --auto-remove \
   -o APT::AutoRemove::RecommendsImportant=false \
   $buildDeps \
   && rm -rf /var/lib/apt/lists/* \
   && rm -rf /tmp/* /var/tmp/* /usr/lib/ruby/gems/*/cache/*.gem

# Copy conf and entrypoint files
COPY fluent.conf /fluentd/etc/
COPY entrypoint.sh /bin/

RUN chmod +x /bin/entrypoint.sh

USER fluent