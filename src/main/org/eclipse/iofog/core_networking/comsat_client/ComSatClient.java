package org.eclipse.iofog.core_networking.comsat_client;

import io.netty.bootstrap.Bootstrap;
import io.netty.channel.Channel;
import io.netty.channel.ChannelInitializer;
import io.netty.channel.EventLoopGroup;
import io.netty.channel.nio.NioEventLoopGroup;
import io.netty.channel.socket.SocketChannel;
import io.netty.channel.socket.nio.NioSocketChannel;
import io.netty.handler.codec.bytes.ByteArrayDecoder;
import io.netty.handler.codec.bytes.ByteArrayEncoder;
import io.netty.handler.ssl.SslContext;
import org.eclipse.iofog.core_networking.main.CoreNetworking;
import org.eclipse.iofog.core_networking.utils.Constants;

import java.util.logging.Logger;

/**
 * Class to connect to ComSat server
 * <p>
 * Created by saeid on 4/8/16.
 */
public class ComSatClient implements Runnable {
    private final Logger log = Logger.getLogger(ComSatClient.class.getName());
    private final SslContext sslCtx;

    private long lastSeen;
    private Channel channel;
    private CoreNetworking coreNetworking = null;
    private int connectionId = -1;
    private boolean disconnect = false;

    public ComSatClient(SslContext sslCtx, CoreNetworking coreNetworking, int connectionId) {
        this.sslCtx = sslCtx;
        this.lastSeen = System.currentTimeMillis();
        this.coreNetworking = coreNetworking;
        this.connectionId = connectionId;
    }

    protected void seen() {
        lastSeen = System.currentTimeMillis();
    }

    private void connect() {
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

            channel = b.connect(CoreNetworking.config.getHost(), CoreNetworking.config.getPort()).sync().channel();
            log.info(String.format("#%d : connected", connectionId));
            coreNetworking.connectingDone();
            try {
                channel.writeAndFlush(CoreNetworking.config.getPassCode().getBytes()).sync();
            } catch (Exception e) {
            }

            channel.closeFuture().sync();
            log.warning(String.format("#%d : disconnected", connectionId));
        } catch (Exception e) {
        	log.warning(String.format("#%d : exception : %s", connectionId, e.getMessage()));
            coreNetworking.connectingDone();
        } finally {
            group.shutdownGracefully();
        }
    }

    public void run() {
        while (true) {
            try {
                log.info(String.format("#%d : connecting...", connectionId));
                connect();
                if (disconnect)
                    break;
                Thread.sleep(100);
                while (coreNetworking.isConnecting()) {
                    Thread.sleep(100);
                }
                coreNetworking.connecting();
            } catch (Exception e) {
            }
        }
    }

    protected int getConnectionId() {
        return connectionId;
    }

    /**
     * sends "BEAT" ASCII to ComSat server
     */
    public void beat() {
        try {
            if (channel == null || !channel.isActive() || System.currentTimeMillis() > (lastSeen + CoreNetworking.config.getHeartbeatThreshold())) {
                log.info(String.format("#%d : CONNECTION FOUND DEAD!!!", connectionId));
                close(false);
            }
            channel.writeAndFlush(Constants.BEAT).sync();
        } catch (Exception e) {
        }
    }

    /**
     * closes the connection
     */
    public void close(boolean disconnect) {
        try {
            this.disconnect = disconnect;
            if (channel != null)
                channel.close();
        } catch (Exception e) {
        }
    }

    /**
     * returns connected channel to ComSat server
     *
     * @return @{@link Channel}
     */
    public Channel getChannel() {
        return this.channel;
    }
}
