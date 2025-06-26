FROM itzg/minecraft-server

# Forge サーバー用の環境変数設定
ENV EULA="true"
ENV TYPE="FORGE"
ENV VERSION="1.20.1"
ENV MINECRAFT_VERSION="1.20.1"
ENV FORGE_VERSION="40.2.0"
ENV MEMORY="4G"
ENV TZ="Asia/Tokyo"

# RCONを有効にする
ENV ENABLE_RCON="true"
ENV RCON_PASSWORD="minecraft"

# JVM オプション (必要に応じてコメントアウトまたは変更)
# ENV JVM_OPTS="-XX:+UseG1GC -XX:+UnlockExperimentalVMOptions"

# デバッグ用JVMオプション (開発時のみ、本番では削除推奨)
# ENV JVM_OPTS="-agentlib:jdwp=transport=dt_socket,server=y,suspend=n,address=5005"

# Forge固有の設定
ENV OVERRIDE_SERVER_PROPERTIES="true"
