package urknall

import "testing"

func TestExtractWriteFile(t *testing.T) {
	s := "mkdir -p /etc/systemd/system \u0026\u0026 echo H4sIAAAJbogA/4yQTUvEMBCG7/kVwx68JXE/sLiQBUUP3mQX8bD0kE1HDaZpnUyKwv54ayu4SAUvwzC8D+/D7B+i51LcYHLkW/ZNNFusfBJbfMueMJmqca9IKiF13qG4emKk30ex341bKW7f0e3YEt8TGqlzIh0aZ4M++KhHDBI3LdBQ85841VPhIXGw6QWkg9k06WNq0fGIr5fqXK1gA7rCTsccAiw2Z3M4HmGabr8iJ+jsR8D8YZojSBltjWbgQLYwLxZqXqhVP9cXy+JyGCC7XsOy/Sb1WDOcTiv7z97FxDaEUjzayFhdf5g6B/Yy989XvcszsvgEAAD//wEAAP//y42HUsYBAAA= | base64 -d | gunzip \u003e /tmp/wunderscale.e125d5b80e2dfaafb2ae44f220572629b2b2e6bc7aeff2de2c6d7bb57e5de635 \u0026\u0026 chown root /tmp/wunderscale.e125d5b80e2dfaafb2ae44f220572629b2b2e6bc7aeff2de2c6d7bb57e5de635 \u0026\u0026 chmod 644 /tmp/wunderscale.e125d5b80e2dfaafb2ae44f220572629b2b2e6bc7aeff2de2c6d7bb57e5de635 \u0026\u0026 mv /tmp/wunderscale.e125d5b80e2dfaafb2ae44f220572629b2b2e6bc7aeff2de2c6d7bb57e5de635 /etc/systemd/system/redis.service"

	path, content, ok, err := extractWriteFile(s)
	if err != nil {
		t.Fatal(err)
	} else if !ok {
		t.Errorf("expected to extract file")
	}
	if v, ex := path, "/etc/systemd/system/redis.service"; ex != v {
		t.Errorf("expected path to be %q, was %q", ex, v)
	}
	et := "[Unit]\nDescription=Redis\nRequires=docker.service\nAfter=docker.service\n\n[Service]\nExecStartPre=-/usr/local/bin/docker stop redis\nExecStartPre=-/usr/local/bin/docker rm redis\nExecStartPre=/bin/bash -c \"/usr/local/bin/docker inspect redis:3.0.4 > /dev/null 2>&1 || /usr/local/bin/docker pull redis:3.0.4\"\nExecStart=/usr/local/bin/docker run --name=redis -p 172.17.42.1:6379:6379 -v /data/docker/redis:/data redis:3.0.4\n\n[Install]\nWantedBy=multi-user.target\n"
	if v, ex := content, et; ex != v {
		t.Errorf("expected content to be %q, was %q", ex, v)
	}
	_ = content
}
