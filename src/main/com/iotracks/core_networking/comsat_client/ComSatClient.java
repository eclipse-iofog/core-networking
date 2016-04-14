package main.com.iotracks.core_networking.comsat_client;

import io.netty.bootstrap.Bootstrap;
import io.netty.channel.Channel;
import io.netty.channel.ChannelFuture;
import io.netty.channel.ChannelInitializer;
import io.netty.channel.EventLoopGroup;
import io.netty.channel.nio.NioEventLoopGroup;
import io.netty.channel.socket.SocketChannel;
import io.netty.channel.socket.nio.NioSocketChannel;
import io.netty.handler.codec.bytes.ByteArrayDecoder;
import io.netty.handler.codec.bytes.ByteArrayEncoder;
import io.netty.handler.ssl.SslContext;
import io.netty.handler.ssl.SslContextBuilder;
import main.com.iotracks.core_networking.utils.Certificate;
import main.com.iotracks.core_networking.main.CoreNetworking;

import java.util.logging.Logger;

/**
 * Class to establish connection to ComSat server
 *
 * Created by saeid on 4/8/16.
 */
public class ComSatClient implements Runnable {
    private final byte[] BEAT = "BEAT".getBytes();
    private final Logger log = Logger.getLogger(ComSatClient.class.getName());
    private final Certificate certificate;

    private SslContext sslCtx;
    private long lastSeen;
    private Channel ch;
    private CoreNetworking coreNetworking = null;
    private int connectionId = -1;
    private boolean disconnect = false;

    public ComSatClient(Certificate certificate, CoreNetworking coreNetworking, int connectionId) {
        this.certificate = certificate;
        this.lastSeen = System.currentTimeMillis();
        this.coreNetworking = coreNetworking;
        this.connectionId = connectionId;
    }

    protected void seen() {
        lastSeen = System.currentTimeMillis();
    }

    private void connect() {
        try {
            sslCtx = SslContextBuilder
                    .forClient()
                    .trustManager(certificate.getCertificate())
                    .build();
        } catch (Exception e) {
            log.warning(String.format("#%d : %s", connectionId, e.getMessage()));
            return;
        }

        EventLoopGroup group = new NioEventLoopGroup(1);
        try {
            ComSatClientHandler handler = new ComSatClientHandler(this);
            Bootstrap b = new Bootstrap();
            b.group(group)
                    .channel(NioSocketChannel.class)
                    .handler(new ChannelInitializer<SocketChannel>() {
                        @Override
                        protected void initChannel(SocketChannel ch) throws Exception {
                            ch.pipeline().addLast(sslCtx.newHandler(ch.alloc(), CoreNetworking.config.getHost(), CoreNetworking.config.getPort()));
                            ch.pipeline().addLast(new ByteArrayDecoder());
                            ch.pipeline().addLast(new ByteArrayEncoder());
                            ch.pipeline().addLast(handler);
                        }
                    });

            ch = b.connect(CoreNetworking.config.getHost(), CoreNetworking.config.getPort()).sync().channel();
            coreNetworking.connectingDone();
            ChannelFuture future = ch.writeAndFlush(CoreNetworking.config.getPassCode().getBytes());
            try {
                future.sync();
            } catch (Exception e) {
            }

            ch.closeFuture().sync();
        } catch (Exception e) {
            coreNetworking.connectingDone();
        } finally {
            group.shutdownGracefully();
        }
    }

    public void run() {
        try {
            while (true) {
                connect();
                if (disconnect)
                    break;
                log.warning(String.format("#%d : connection lost. connecting...", connectionId));
                Thread.sleep(1000);
                while (coreNetworking.isConnecting()) {
                    Thread.sleep(100);
                }
                coreNetworking.connecting();
            }
        } catch (Exception e) {
        }
    }

    protected int getConnectionId() {
        return connectionId;
    }

    public void beat() {
        if (System.currentTimeMillis() > (lastSeen + CoreNetworking.config.getHeartbeatThreshold())) {
            ch.eventLoop().shutdownGracefully();
            ch.close();
        } else {
            ChannelFuture future = ch.writeAndFlush(BEAT);
            try {
                future.sync();
            } catch (Exception e) {
            }
        }
    }

    public void close() {
        disconnect = true;
        ch.disconnect();
    }

    public Channel getChannel() {
        return this.ch;
    }
}
