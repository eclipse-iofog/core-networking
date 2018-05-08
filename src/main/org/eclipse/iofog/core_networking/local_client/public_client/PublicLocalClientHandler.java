package org.eclipse.iofog.core_networking.local_client.public_client;

import io.netty.channel.Channel;
import io.netty.channel.ChannelHandlerContext;
import io.netty.channel.SimpleChannelInboundHandler;

import java.util.logging.Logger;

/**
 * local container connection handler
 * <p>
 * Created by saeid on 4/12/16.
 */
public class PublicLocalClientHandler extends SimpleChannelInboundHandler<byte[]> {
    private final Logger log= Logger.getLogger(PublicLocalClientHandler.class.getName());
    private final Channel comSatChannel;

    public PublicLocalClientHandler(Channel comSatChannel) {
        this.comSatChannel = comSatChannel;
    }

    /**
     * disconnects from ComSat server when connection to local container been lost.
     *
     * @param ctx
     * @throws Exception
     */
    @Override
    public void channelInactive(ChannelHandlerContext ctx) throws Exception {
//        if (comSatChannel != null) {
//            try {
//                comSatChannel.disconnect();
//                comSatChannel.close();
//            } catch (Exception e) {
//            }
//        }
    }

    /**
     * pipes received data from local container to ComSat server
     *
     * @param ctx
     * @param bytes
     * @throws Exception
     */
    @Override
    protected void channelRead0(ChannelHandlerContext ctx, byte[] bytes) throws Exception {
        try {
            log.info("piping bytes from client to comsat");
            comSatChannel.writeAndFlush(bytes).sync();
            ctx.disconnect();
        } catch (Exception e) {
        }
    }

    @Override
    public void exceptionCaught(ChannelHandlerContext ctx, Throwable cause) {
        log.warning(String.format("exception in client connection : %s", cause.getMessage()));
        ctx.disconnect();
    }

}
